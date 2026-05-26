package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/PuerkitoBio/goquery"
	"github.com/dop251/goja"
	"github.com/otiai10/gosseract/v2"
	"golang.org/x/image/draw"
)

const (
	BaseURL   = "https://nensaysubs.net"
	UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0"
)

const captchaPreviewWidth = 80
const maxHTTPRetries = 3
const flareSolverrTimeout = 120000
const mediaRootDefault = "/data/media"

var runtimeUserAgent = UserAgent
var runtimeFlareSession = ""

type episode struct {
	Title         string
	EpisodeNumber int
	Filename      string
	DownloadURL   string
	EditURL       string
	PageURL       string
	Token         string
}

func main() {
	args := os.Args[1:]

	// --move-only mode: skip download, run only the post-processing (move/rename) step.
	// Usage: nensaysubs --move-only [subsDir [baseVideoDir]]
	if len(args) > 0 && (args[0] == "--move-only" || args[0] == "-m") {
		dest := strings.TrimSpace(os.Getenv("NENSAYSUBS_OUTPUT_DIR"))
		if len(args) > 1 {
			dest = args[1]
		}
		if dest == "" {
			dest = "/data/media/subs"
		}

		baseDir := strings.TrimSpace(os.Getenv("NENSAYSUBS_EPISODE_BASE_DIR"))
		if len(args) > 2 {
			baseDir = args[2]
		}

		interactive := envBool("NENSAYSUBS_INTERACTIVE", true)

		if err := runMoveOnly(dest, baseDir, interactive); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	series := ""
	if len(args) > 0 {
		series = args[0]
	} else {
		series = strings.TrimSpace(os.Getenv("NENSAYSUBS_SERIES"))
	}

	if series == "" {
		fmt.Fprintln(os.Stderr, "uso: nensaysubs <SerieSlug> [destino]")
		fmt.Fprintln(os.Stderr, "     nensaysubs Maoyuu_Maou_Yuusha /data/media/subs")
		fmt.Fprintln(os.Stderr, "     nensaysubs --move-only [subsDir [baseVideoDir]]")
		os.Exit(1)
	}

	dest := strings.TrimSpace(os.Getenv("NENSAYSUBS_OUTPUT_DIR"))
	if len(args) > 1 {
		dest = args[1]
	}
	if dest == "" {
		dest = "."
	}

	interactive := envBool("NENSAYSUBS_INTERACTIVE", true)

	if err := runSeriesDownload(series, dest, interactive); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func envBool(name string, defaultValue bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if raw == "" {
		return defaultValue
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func runMoveOnly(destination string, baseDir string, interactive bool) error {
	if destination == "" {
		destination = "/data/media/subs"
	}

	fmt.Printf("[+] Modo --move-only. Directorio de subtitulos: %s\n", destination)

	// Extract any leftover ZIPs/RARs that may still be packed
	if extractedCount, zipErr := extractAllPendingZips(destination, 4); zipErr != nil {
		fmt.Printf("[DEBUG] Error extrayendo archivos pendientes: %v\n", zipErr)
	} else if extractedCount > 0 {
		fmt.Printf("[DEBUG] Archivos pendientes extraidos: %d\n", extractedCount)
	}

	srtCount, assCount, ssaCount, scanErr := scanSubtitleCounts(destination)
	if scanErr != nil {
		fmt.Printf("[DEBUG] No se pudo escanear subtitulos en %s: %v\n", destination, scanErr)
	} else {
		fmt.Printf("[+] Subtitulos disponibles en %s: .srt=%d .ass=%d .ssa=%d\n", destination, srtCount, assCount, ssaCount)
	}

	if baseDir == "" && interactive {
		if askYesNo("[?] Quieres mover/renombrar subtitulos usando una carpeta base de episodios? (s/N): ", false) {
			selected, selErr := chooseDirectoryInteractive(mediaRootDefault)
			if selErr != nil {
				return fmt.Errorf("se omitio el post-proceso: %w", selErr)
			}
			baseDir = selected
		}
	}

	if baseDir == "" {
		return fmt.Errorf("se requiere directorio base de episodios (baseVideoDir o NENSAYSUBS_EPISODE_BASE_DIR)")
	}

	manualRequested := false
	if interactive {
		manualRequested = askYesNo("[?] Activar modo manual asistido para peliculas/specials/multiidioma? (s/N): ", true)
	}

	if manualRequested {
		manualMoved, manualRemaining, manualErr := moveSubtitlesManual(destination, baseDir)
		if manualErr != nil {
			fmt.Printf("[!] Error en modo manual: %v\n", manualErr)
		} else {
			fmt.Printf("[+] Modo manual completado. movidos=%d restantes=%d\n", manualMoved, manualRemaining)
		}
	} else {
		moved, skipped, moveErr := moveAndRenameSubtitles(destination, baseDir)
		if moveErr != nil {
			fmt.Printf("[!] Error moviendo subtitulos: %v\n", moveErr)
		} else {
			fmt.Printf("[+] Post-proceso automatico completado. movidos=%d omitidos=%d\n", moved, skipped)
		}

		remaining, remErr := listSubtitleFiles(destination)
		if remErr != nil {
			fmt.Printf("[DEBUG] No se pudo listar subtitulos restantes: %v\n", remErr)
		} else if len(remaining) > 0 {
			multiLang := hasMultipleLanguageHints(remaining)
			if interactive && (skipped > 0 || multiLang) {
				msg := "[?] Se detectaron subtitulos restantes"
				if multiLang {
					msg += " (posible multiidioma)"
				}
				msg += ". Quieres abrir modo manual asistido? (S/n): "
				if askYesNo(msg, false) {
					manualMoved, manualRemaining, manualErr := moveSubtitlesManual(destination, baseDir)
					if manualErr != nil {
						fmt.Printf("[!] Error en modo manual: %v\n", manualErr)
					} else {
						fmt.Printf("[+] Modo manual completado. movidos=%d restantes=%d\n", manualMoved, manualRemaining)
					}
				}
			}
		}
	}

	return nil
}

func runSeriesDownload(series string, destination string, interactive bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during download: %v", r)
		}
	}()

	defer func() {
		if cleanErr := cleanupSubsWorkspace("/data/media/subs"); cleanErr != nil {
			fmt.Printf("[!] Limpieza final fallo en /data/media/subs: %v\n", cleanErr)
		} else {
			fmt.Println("[+] Limpieza final completada: /data/media/subs vacio")
		}
	}()

	if strings.TrimSpace(series) == "" {
		return fmt.Errorf("series is required")
	}

	if strings.TrimSpace(destination) == "" {
		destination = "."
	}

	seriesURL := BaseURL + "/sub/" + series

	jar, _ := cookiejar.New(nil)

	client := newHTTPClient(jar)

	client.CheckRedirect = func(
		req *http.Request,
		via []*http.Request,
	) error {

		return http.ErrUseLastResponse
	}

	defer func() {
		if runtimeFlareSession != "" {
			if err := destroyFlareSolverrSession(client); err != nil {
				fmt.Printf("[DEBUG] No se pudo cerrar sesión FlareSolverr: %v\n", err)
			}
		}
	}()

	fmt.Println("[DEBUG] HTTP transport mode: h1-only (HTTP/2 disabled for VPN stability)")

	fmt.Println("[+] Opening initial series page...")
	fmt.Println(seriesURL)

	resp, err := doGetWithRetry(client, seriesURL, "")
	check(err)

	defer resp.Body.Close()

	fmt.Println("[DEBUG] Initial status:", resp.Status)

	if loc := resp.Header.Get("Location"); loc != "" {

		fmt.Println("[DEBUG] Redirect:", loc)
	}

	// solve entry gate
	if resp.StatusCode == 302 {

		err = solveEntryGate(client)

		check(err)
	}

	fmt.Println("[+] Re-opening series page after auth...")

	resp, err = doGetWithRetry(client, seriesURL, BaseURL+"/ingreso/")
	check(err)

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	check(err)

	body := string(bodyBytes)

	fmt.Println("[DEBUG] Final status:", resp.Status)
	fmt.Println("[DEBUG] Final URL:", resp.Request.URL.String())

	if isVerificationPage(body) {
		fmt.Println("[!] Página anti-bot detectada. Intentando resolver con FlareSolverr...")
		flareBody, err := solveWithFlareSolverr(client, seriesURL)
		if err != nil {
			fmt.Println(body)
			return fmt.Errorf("anti-bot detectado y no se pudo resolver: %w", err)
		}

		resp2, err := doGetWithRetry(client, seriesURL, BaseURL+"/")
		check(err)
		defer resp2.Body.Close()

		bodyBytes2, err := io.ReadAll(resp2.Body)
		check(err)
		body = string(bodyBytes2)

		fmt.Println("[DEBUG] Final status tras FlareSolverr:", resp2.Status)
		fmt.Println("[DEBUG] Final URL tras FlareSolverr:", resp2.Request.URL.String())

		if isVerificationPage(body) && strings.TrimSpace(flareBody) != "" {
			fmt.Println("[DEBUG] Re-fetch aún bloqueado; usando HTML resuelto por FlareSolverr.")
			body = flareBody
		}
	}

	if isIngresoGatePage(body) {
		fmt.Println("[!] Gate de ingreso detectado. Resolviendo captcha matemático...")
		if err := solveEntryGate(client); err != nil {
			return fmt.Errorf("no se pudo resolver gate de ingreso: %w", err)
		}

		resp3, err := doGetWithRetry(client, seriesURL, BaseURL+"/ingreso/")
		check(err)
		defer resp3.Body.Close()

		bodyBytes3, err := io.ReadAll(resp3.Body)
		check(err)
		body = string(bodyBytes3)

		fmt.Println("[DEBUG] Status tras resolver gate:", resp3.Status)
		fmt.Println("[DEBUG] URL tras resolver gate:", resp3.Request.URL.String())
	}

	if isVerificationPage(body) {
		fmt.Println("[!] Challenge JS custom detectado. Intentando emulación local del navegador...")
		solved, newBody, err := solveCustomVerificationChallenge(client, body, seriesURL)
		if err != nil {
			fmt.Printf("[DEBUG] Emulación challenge falló: %v\n", err)
		}
		if solved {
			body = newBody
			fmt.Println("[+] Challenge JS superado con emulación local.")
		}
	}

	pages, pageErr := collectSeriesPages(client, seriesURL, body)
	if pageErr != nil {
		return fmt.Errorf("no se pudo recorrer paginacion de serie: %w", pageErr)
	}

	episodes := parseEpisodesFromPages(pages)
	if len(episodes) == 0 {
		fmt.Println(body)
		return fmt.Errorf("no se encontraron episodios en la página")
	}

	if interactive {
		episodes = selectEpisodes(episodes)
	}
	if len(episodes) == 0 {
		return fmt.Errorf("ningún episodio seleccionado")
	}

	fmt.Printf("[+] Descargando %d episodio(s)...\n", len(episodes))

	downloadedCount := 0
	for idx, ep := range episodes {
		displayName := strings.TrimSpace(ep.Title)
		if displayName == "" {
			displayName = strings.TrimSpace(ep.Filename)
		}
		if displayName == "" {
			displayName = fmt.Sprintf("Episodio %02d", idx+1)
		}

		fmt.Printf("\n[+] Procesando %d/%d: %s\n", idx+1, len(episodes), displayName)

		episodeURL := strings.TrimSpace(ep.DownloadURL)
		if episodeURL == "" && ep.Token != "" {
			episodeURL = BaseURL + "/senos/" + ep.Token
		}
		if episodeURL == "" {
			fmt.Printf("[!] Episodio sin link de descarga. Omitiendo: %s\n", displayName)
			continue
		}

		fmt.Println("[+] Opening episode page...")
		fmt.Println(episodeURL)

		success := false
		var zipData []byte

		for attempt := 1; attempt <= 20; attempt++ {

			fmt.Printf(
				"\n[+] Captcha intento %d/20 (episodio %d/%d)\n",
				attempt,
				idx+1,
				len(episodes),
			)

			// open captcha page fresh every attempt
			epResp, err := doGetWithRetry(client, episodeURL, seriesURL)
			check(err)

			epBytes, err := io.ReadAll(epResp.Body)
			epResp.Body.Close()
			check(err)

			epHTML := string(epBytes)

			reField := regexp.MustCompile(
				`name="([a-zA-Z0-9_]+)"`,
			)

			fieldMatches := reField.FindAllStringSubmatch(
				epHTML,
				-1,
			)

			postField := "code"

			fmt.Println("[DEBUG] Searching captcha form field...")
			for _, m := range fieldMatches {
				fmt.Println("[DEBUG] Found input field:", m[1])
				if m[1] != "submit" {
					postField = m[1]
				}
			}

			fmt.Println("[+] Using POST field:", postField)

			captchaResp, err := doGetWithRetry(client, BaseURL+"/senos/seguro.php", episodeURL)
			check(err)

			captchaBytes, err := io.ReadAll(captchaResp.Body)
			captchaResp.Body.Close()
			check(err)

			err = os.WriteFile("captcha_raw.png", captchaBytes, 0644)
			check(err)

			err = cleanCaptcha("captcha_raw.png", "captcha_clean.png")
			check(err)

			if interactive {
				fmt.Println("[+] Captcha preview (terminal):")
				if err := printImageASCII("captcha_clean.png", captchaPreviewWidth); err != nil {
					fmt.Println("[DEBUG] Could not print captcha preview:", err)
				}
			}

			text, err := solveCaptchaAuto("captcha_raw.png")
			if err != nil {
				fmt.Println("[DEBUG] OCR failed:", err)
				text = ""
			}

			if interactive {
				var useIt bool
				text, useIt = askCaptchaValue(text)
				if !useIt {
					os.Remove("captcha_raw.png")
					os.Remove("captcha_clean.png")
					cleanupCaptchaVariants()
					continue
				}
			} else {
				fmt.Println("[+] Captcha OCR:", text)
			}

			fmt.Println("[+] Captcha a enviar:", text)

			if text == "" {
				fmt.Println("[DEBUG] empty captcha")
				os.Remove("captcha_raw.png")
				os.Remove("captcha_clean.png")
				cleanupCaptchaVariants()
				continue
			}

			fmt.Println("[+] Sending captcha:", text)

			form := url.Values{}
			form.Set(postField, text)

			req, err := http.NewRequest(
				"POST",
				BaseURL+"/solicitud/",
				strings.NewReader(form.Encode()),
			)
			check(err)

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Referer", episodeURL)
			req.Header.Set("User-Agent", runtimeUserAgent)

			downloadResp, err := client.Do(req)
			check(err)

			contentType := downloadResp.Header.Get("Content-Type")
			location := downloadResp.Header.Get("Location")

			if location != "" {
				fmt.Println("[DEBUG] Redirecting to:", location)
			}

			if strings.Contains(contentType, "application/zip") ||
				strings.Contains(contentType, "application/octet-stream") {

				fmt.Println("[+] ZIP downloaded directly")
				zipData, err = io.ReadAll(downloadResp.Body)
				downloadResp.Body.Close()
				check(err)
				success = true
				break
			}

			if location != "" {
				baseRespURL := downloadResp.Request.URL
				nextURL, err := baseRespURL.Parse(location)
				check(err)

				finalResp, err := doGetWithRetry(client, nextURL.String(), BaseURL+"/solicitud/")
				check(err)

				finalType := finalResp.Header.Get("Content-Type")
				if strings.Contains(finalType, "application/zip") ||
					strings.Contains(finalType, "application/octet-stream") {

					fmt.Println("[+] ZIP downloaded after redirect")
					zipData, err = io.ReadAll(finalResp.Body)
					finalResp.Body.Close()
					check(err)
					success = true
					break
				}

				finalBytes, _ := io.ReadAll(finalResp.Body)
				finalResp.Body.Close()
				if len(finalBytes) > 0 {
					snippet := string(finalBytes)
					if len(snippet) > 220 {
						snippet = snippet[:220]
					}
					fmt.Println("[DEBUG] Final non-zip snippet:")
					fmt.Println(snippet)
				}
			}

			if location == "" {
				respBytes, _ := io.ReadAll(downloadResp.Body)
				if len(respBytes) > 0 {
					snippet := string(respBytes)
					if len(snippet) > 220 {
						snippet = snippet[:220]
					}
					fmt.Println("[DEBUG] Non-zip snippet:")
					fmt.Println(snippet)
				}
			}

			downloadResp.Body.Close()

			fmt.Println("")
			fmt.Println("===================================")
			fmt.Println("Captcha incorrecto.")
			fmt.Println("Nuevo captcha generado.")
			fmt.Println("===================================")

			os.Remove("captcha_raw.png")
			os.Remove("captcha_clean.png")
			cleanupCaptchaVariants()
		}

		os.Remove("captcha_raw.png")
		os.Remove("captcha_clean.png")
		cleanupCaptchaVariants()

		if !success {
			fmt.Println("[!] Could not download this token, continuing...")
			continue
		}

		fmt.Println("[+] Extrayendo pack...")

		err = unzip(zipData, destination)
		check(err)

		if nestedCount, nestedErr := extractAllPendingZips(destination, 4); nestedErr != nil {
			fmt.Printf("[DEBUG] Error extrayendo ZIP anidados: %v\n", nestedErr)
		} else if nestedCount > 0 {
			fmt.Printf("[DEBUG] ZIP anidados extraidos tras descarga: %d\n", nestedCount)
		}

		downloadedCount++
		fmt.Println("[+] Episodio completado.")
	}

	if downloadedCount == 0 {
		return fmt.Errorf("no se pudo descargar ningún episodio")
	}

	fmt.Printf("[+] Listo. Descargados %d/%d episodios.\n", downloadedCount, len(episodes))

	baseDir := strings.TrimSpace(os.Getenv("NENSAYSUBS_EPISODE_BASE_DIR"))
	if interactive {
		if askYesNo("[?] Quieres mover/renombrar SRT usando una carpeta base de episodios? (s/N): ", false) {
			selected, selErr := chooseDirectoryInteractive(mediaRootDefault)
			if selErr != nil {
				fmt.Printf("[!] Se omitio el post-proceso: %v\n", selErr)
			} else {
				baseDir = selected
			}
		}
	}

	if baseDir != "" {
		extractedZipCount, zipErr := extractPendingZips(destination)
		if zipErr != nil {
			fmt.Printf("[DEBUG] No se pudieron procesar ZIP pendientes en %s: %v\n", destination, zipErr)
		} else if extractedZipCount > 0 {
			fmt.Printf("[DEBUG] ZIP pendientes extraidos antes del post-proceso: %d\n", extractedZipCount)
		}

		srtCount, assCount, ssaCount, scanErr := scanSubtitleCounts(destination)
		if scanErr != nil {
			fmt.Printf("[DEBUG] No se pudo escanear subtitulos en %s: %v\n", destination, scanErr)
		} else {
			fmt.Printf("[DEBUG] Pre post-proceso en %s: .srt=%d .ass=%d .ssa=%d\n", destination, srtCount, assCount, ssaCount)
		}

		finalSrtCount, finalAssCount, finalSsaCount, finalScanErr := scanSubtitleCounts(destination)
		if finalScanErr == nil {
			fmt.Printf("[DEBUG] Subtitulos detectados para mover: .srt=%d .ass=%d .ssa=%d\n", finalSrtCount, finalAssCount, finalSsaCount)
		}

		manualRequested := false
		if interactive {
			manualRequested = askYesNo("[?] Activar modo manual asistido para peliculas/specials/multiidioma? (s/N): ", true)
		}

		if manualRequested {
			manualMoved, manualRemaining, manualErr := moveSubtitlesManual(destination, baseDir)
			if manualErr != nil {
				fmt.Printf("[!] Error en modo manual: %v\n", manualErr)
			} else {
				fmt.Printf("[+] Modo manual completado. movidos=%d restantes=%d\n", manualMoved, manualRemaining)
			}
		} else {
			moved, skipped, moveErr := moveAndRenameSubtitles(destination, baseDir)
			if moveErr != nil {
				fmt.Printf("[!] Error moviendo subtitulos: %v\n", moveErr)
			} else {
				fmt.Printf("[+] Post-proceso SRT automatico completado. movidos=%d omitidos=%d\n", moved, skipped)
			}

			remaining, remErr := listSubtitleFiles(destination)
			if remErr != nil {
				fmt.Printf("[DEBUG] No se pudo listar subtitulos restantes: %v\n", remErr)
			} else if len(remaining) > 0 {
				multiLang := hasMultipleLanguageHints(remaining)
				if interactive && (skipped > 0 || multiLang) {
					msg := "[?] Se detectaron subtitulos restantes"
					if multiLang {
						msg += " (posible multiidioma)"
					}
					msg += ". Quieres abrir modo manual asistido? (S/n): "
					if askYesNo(msg, false) {
						manualMoved, manualRemaining, manualErr := moveSubtitlesManual(destination, baseDir)
						if manualErr != nil {
							fmt.Printf("[!] Error en modo manual: %v\n", manualErr)
						} else {
							fmt.Printf("[+] Modo manual completado. movidos=%d restantes=%d\n", manualMoved, manualRemaining)
						}
					}
				}
			}
		}
	}

	if hook := strings.TrimSpace(os.Getenv("NENSAYSUBS_POST_SCRIPT")); hook != "" {
		if err := runPostScript(hook, destination, series); err != nil {
			fmt.Printf("[!] Post-script fallo: %v\n", err)
		}
	}

	return nil
}

func cleanCaptcha(
	input string,
	output string,
) error {
	return cleanCaptchaAdvanced(input, output, 180, 4, false)
}

func cleanCaptchaAdvanced(
	input string,
	output string,
	threshold uint8,
	scale int,
	invert bool,
) error {

	in, err := os.Open(input)

	if err != nil {
		return err
	}

	defer in.Close()

	img, _, err := image.Decode(in)

	if err != nil {
		return err
	}

	bounds := img.Bounds()

	// Captcha text is usually centered; trimming side margins removes noisy edges.
	trimX := bounds.Dx() / 10
	trimY := bounds.Dy() / 10
	if trimX > 0 && trimY > 0 {
		bounds = image.Rect(
			bounds.Min.X+trimX,
			bounds.Min.Y+trimY,
			bounds.Max.X-trimX,
			bounds.Max.Y-trimY,
		)
	}

	gray := image.NewGray(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {

		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			r, g, b, _ := img.At(x, y).RGBA()

			avg := uint8(
				((r >> 8) + (g >> 8) + (b >> 8)) / 3,
			)

			isWhite := avg > threshold
			if invert {
				isWhite = !isWhite
			}

			gx := x - bounds.Min.X
			gy := y - bounds.Min.Y
			if isWhite {
				gray.SetGray(gx, gy, color.Gray{255})
			} else {
				gray.SetGray(gx, gy, color.Gray{0})
			}
		}
	}

	// Keep only likely character components and trim empty margins.
	gray = filterCharacterComponents(gray)
	gray = cropToForeground(gray, 2)

	// upscale after denoise for easier OCR segmentation.
	scaled := image.NewGray(
		image.Rect(
			0,
			0,
			gray.Bounds().Dx()*scale,
			gray.Bounds().Dy()*scale,
		),
	)

	draw.NearestNeighbor.Scale(
		scaled,
		scaled.Bounds(),
		gray,
		gray.Bounds(),
		draw.Over,
		nil,
	)

	out, err := os.Create(output)

	if err != nil {
		return err
	}

	defer out.Close()

	return png.Encode(out, scaled)
}

func filterCharacterComponents(src *image.Gray) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.SetGray(x, y, color.Gray{255})
		}
	}

	w := b.Dx()
	h := b.Dy()
	if w == 0 || h == 0 {
		return dst
	}

	visited := make([]bool, w*h)
	idx := func(x, y int) int {
		return (y-b.Min.Y)*w + (x - b.Min.X)
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if src.GrayAt(x, y).Y != 0 || visited[idx(x, y)] {
				continue
			}

			queue := []image.Point{{X: x, Y: y}}
			visited[idx(x, y)] = true
			pixels := make([]image.Point, 0, 128)

			minX, maxX := x, x
			minY, maxY := y, y

			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				pixels = append(pixels, p)

				if p.X < minX {
					minX = p.X
				}
				if p.X > maxX {
					maxX = p.X
				}
				if p.Y < minY {
					minY = p.Y
				}
				if p.Y > maxY {
					maxY = p.Y
				}

				for ny := p.Y - 1; ny <= p.Y+1; ny++ {
					for nx := p.X - 1; nx <= p.X+1; nx++ {
						if nx < b.Min.X || nx >= b.Max.X || ny < b.Min.Y || ny >= b.Max.Y {
							continue
						}
						ii := idx(nx, ny)
						if visited[ii] || src.GrayAt(nx, ny).Y != 0 {
							continue
						}
						visited[ii] = true
						queue = append(queue, image.Point{X: nx, Y: ny})
					}
				}
			}

			area := len(pixels)
			compW := maxX - minX + 1
			compH := maxY - minY + 1
			density := float64(area) / float64(compW*compH)

			// Heuristics tuned for this captcha: keep compact letter/number blobs.
			keep := true
			if area < 18 {
				keep = false
			}
			if compH < 6 || compW < 2 {
				keep = false
			}
			if (compW > 30 || compH > 26) && density < 0.20 {
				keep = false
			}
			if density < 0.08 {
				keep = false
			}

			if keep {
				for _, p := range pixels {
					dst.SetGray(p.X, p.Y, color.Gray{0})
				}
			}
		}
	}

	return dst
}

func cropToForeground(src *image.Gray, pad int) *image.Gray {
	b := src.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	found := false

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if src.GrayAt(x, y).Y != 0 {
				continue
			}
			found = true
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	if !found {
		return src
	}

	minX -= pad
	minY -= pad
	maxX += pad
	maxY += pad

	if minX < b.Min.X {
		minX = b.Min.X
	}
	if minY < b.Min.Y {
		minY = b.Min.Y
	}
	if maxX >= b.Max.X {
		maxX = b.Max.X - 1
	}
	if maxY >= b.Max.Y {
		maxY = b.Max.Y - 1
	}

	cb := image.Rect(minX, minY, maxX+1, maxY+1)
	out := image.NewGray(image.Rect(0, 0, cb.Dx(), cb.Dy()))
	for y := 0; y < cb.Dy(); y++ {
		for x := 0; x < cb.Dx(); x++ {
			out.SetGray(x, y, src.GrayAt(cb.Min.X+x, cb.Min.Y+y))
		}
	}

	return out
}

func solveCaptchaAuto(rawPath string) (string, error) {
	type variant struct {
		threshold uint8
		scale     int
		invert    bool
	}

	variants := []variant{
		{threshold: 140, scale: 5, invert: false},
		{threshold: 160, scale: 5, invert: false},
		{threshold: 180, scale: 5, invert: false},
		{threshold: 170, scale: 5, invert: false},
		{threshold: 170, scale: 5, invert: true},
	}

	modes := []gosseract.PageSegMode{
		gosseract.PSM_SINGLE_WORD,
		gosseract.PSM_SINGLE_LINE,
		gosseract.PSM_RAW_LINE,
	}

	best := ""
	bestScore := -1

	for i, v := range variants {
		out := "captcha_variant_" + strconv.Itoa(i+1) + ".png"
		err := cleanCaptchaAdvanced(rawPath, out, v.threshold, v.scale, v.invert)
		if err != nil {
			continue
		}

		for _, mode := range modes {
			ocr := gosseract.NewClient()
			ocr.SetImage(out)
			ocr.SetWhitelist("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
			_ = ocr.SetVariable("tessedit_char_blacklist", "_-=+~!@#$%^&*()[]{}|;:'\",.<>?/\\")
			ocr.SetPageSegMode(mode)

			text, err := ocr.Text()
			ocr.Close()
			if err != nil {
				continue
			}

			text = sanitizeCaptchaText(text)
			text = reduceCandidate(text)

			score := captchaScore(text)
			if score > bestScore {
				bestScore = score
				best = text
			}
			if isLikelyCaptcha(text) && score >= 80 {
				return text, nil
			}
		}
	}

	if best == "" {
		return "", fmt.Errorf("empty OCR result")
	}

	return best, nil
}

func reduceCandidate(text string) string {
	if len(text) <= 5 {
		return text
	}

	// Prefer windows that contain both letters and digits (common in this captcha).
	best := text[:5]
	bestScore := captchaScore(best)
	for win := 3; win <= 5; win++ {
		if len(text) < win {
			continue
		}
		for i := 0; i+win <= len(text); i++ {
			s := text[i : i+win]
			sc := captchaScore(s)
			if sc > bestScore {
				bestScore = sc
				best = s
			}
		}
	}

	return best
}

func sanitizeCaptchaText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, " ", "")
	return regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(text, "")
}

func askCaptchaValue(guess string) (string, bool) {
	reader := bufio.NewReader(os.Stdin)

	if guess != "" {
		fmt.Printf("[?] OCR sugiere: %s\n", guess)
		fmt.Print("    Enter=usar, escribe correccion, 0=reintentar: ")
	} else {
		fmt.Print("[?] OCR vacio. Escribe captcha manual o 0 para reintentar: ")
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		line = strings.TrimSpace(line)
	}

	input := sanitizeCaptchaText(line)
	if strings.TrimSpace(line) == "0" {
		return "", false
	}

	if input == "" {
		if guess == "" {
			return "", false
		}
		return guess, true
	}

	return input, true
}

func isLikelyCaptcha(text string) bool {
	// Captcha observed on site is usually short alphanumeric.
	return len(text) >= 3 && len(text) <= 5
}

func captchaScore(text string) int {
	if text == "" {
		return 0
	}

	score := 0
	if len(text) >= 3 && len(text) <= 5 {
		score += 60
	} else {
		score -= 20
	}

	hasDigit := false
	hasLetter := false

	repeatPenalty := 0
	for i, r := range text {
		if r >= '0' && r <= '9' {
			hasDigit = true
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		}
		if i > 0 && rune(text[i-1]) == r {
			repeatPenalty += 5
		}
	}

	if hasDigit {
		score += 15
	}
	if hasLetter {
		score += 15
	}
	if hasDigit && hasLetter {
		score += 20
	}

	score -= repeatPenalty

	if score < 0 {
		return 0
	}
	return score
}

func cleanupCaptchaVariants() {
	for i := 1; i <= 5; i++ {
		os.Remove("captcha_variant_" + strconv.Itoa(i) + ".png")
	}
}

func printImageASCII(path string, targetWidth int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w == 0 || h == 0 {
		return fmt.Errorf("empty image")
	}

	if targetWidth <= 0 {
		targetWidth = 80
	}

	// Terminal characters are taller than wide, so compress height by ~2.
	targetHeight := (h * targetWidth) / (w * 2)
	if targetHeight < 1 {
		targetHeight = 1
	}

	gradient := " .:-=+*#%@"
	last := len(gradient) - 1

	for y := 0; y < targetHeight; y++ {
		var line strings.Builder
		line.Grow(targetWidth)

		sy := b.Min.Y + (y*h)/targetHeight
		for x := 0; x < targetWidth; x++ {
			sx := b.Min.X + (x*w)/targetWidth
			r, g, bl, _ := img.At(sx, sy).RGBA()
			gray := uint8(((r >> 8) + (g >> 8) + (bl >> 8)) / 3)

			idx := int(gray) * last / 255
			line.WriteByte(gradient[idx])
		}

		fmt.Println(line.String())
	}

	return nil
}

func solveEntryGate(
	client *http.Client,
) error {

	fmt.Println("[+] Solving entry gate...")
	fmt.Println("[+] Opening gate...")

	resp, err := doGetWithRetry(client, BaseURL+"/ingreso/", BaseURL+"/")

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	body := string(bodyBytes)
	if isVerificationPage(body) {
		fmt.Println("[DEBUG] Gate request cayó en anti-bot. Intentando FlareSolverr sobre /ingreso/...")
		flareBody, fsErr := solveWithFlareSolverr(client, BaseURL+"/ingreso/")
		if fsErr != nil {
			return fmt.Errorf("gate bloqueado por anti-bot: %w", fsErr)
		}

		resp2, err := doGetWithRetry(client, BaseURL+"/ingreso/", BaseURL+"/")
		if err == nil {
			defer resp2.Body.Close()
			body2, readErr := io.ReadAll(resp2.Body)
			if readErr == nil {
				body = string(body2)
				fmt.Println("[DEBUG] Gate status tras FlareSolverr:", resp2.Status)
			}
		}

		if isVerificationPage(body) && strings.TrimSpace(flareBody) != "" {
			fmt.Println("[DEBUG] Gate sigue bloqueado tras re-fetch; usando HTML de FlareSolverr.")
			body = flareBody
		}
	}

	fmt.Println("[DEBUG] Gate status:", resp.Status)

	snippetLen := 2500

	if len(body) < snippetLen {

		snippetLen = len(body)
	}

	fmt.Println(
		"[DEBUG] Gate snippet:\n" +
			body[:snippetLen],
	)

	reMath := regexp.MustCompile(
		`([0-9]+)\s*\+\s*([0-9]+)`,
	)

	match := reMath.FindStringSubmatch(body)

	if len(match) < 3 {

		return fmt.Errorf(
			"math expression not found",
		)
	}

	a := atoi(match[1])
	b := atoi(match[2])

	result := a + b

	fmt.Printf(
		"[+] Solved: %d + %d = %d\n",
		a,
		b,
		result,
	)

	form := url.Values{}

	form.Set(
		"valor",
		fmt.Sprintf("%d", result),
	)

	postResp, err := doPostFormWithRetry(client, BaseURL+"/ingreso/index.php", form, BaseURL+"/ingreso/")

	if err != nil {
		return err
	}

	defer postResp.Body.Close()

	fmt.Println(
		"[DEBUG] Gate POST status:",
		postResp.Status,
	)

	if loc := postResp.Header.Get("Location"); loc != "" {

		fmt.Println(
			"[DEBUG] Gate redirect:",
			loc,
		)
	}

	u, _ := url.Parse(BaseURL)

	fmt.Println("[DEBUG] Cookies:")

	for _, c := range client.Jar.Cookies(u) {

		fmt.Printf(
			"  %s=%s\n",
			c.Name,
			c.Value,
		)
	}

	return nil
}

func unzip(
	data []byte,
	dest string,
) error {

	reader, err := zip.NewReader(
		bytes.NewReader(data),
		int64(len(data)),
	)

	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, 0755)

	if err != nil {
		return err
	}

	for _, file := range reader.File {

		path := dest + "/" + file.Name

		if file.FileInfo().IsDir() {

			os.MkdirAll(path, file.Mode())

			continue
		}

		rc, err := file.Open()

		if err != nil {
			return err
		}

		out, err := os.Create(path)

		if err != nil {

			rc.Close()

			return err
		}

		_, err = io.Copy(out, rc)

		rc.Close()
		out.Close()

		if err != nil {
			return err
		}

		fmt.Printf(
			"[+] Extracted: %s\n",
			path,
		)
	}

	return nil
}

func extractArchiveFile(path string, dest string) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".zip":
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return unzip(data, dest)
	case ".rar":
		if _, err := exec.LookPath("bsdtar"); err != nil {
			return fmt.Errorf("bsdtar no disponible para extraer RAR")
		}
		cmd := exec.Command("bsdtar", "-xf", path, "-C", dest)
		var stderr bytes.Buffer
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			msg := strings.TrimSpace(stderr.String())
			if msg == "" {
				msg = err.Error()
			}
			return fmt.Errorf("bsdtar fallo extrayendo %s: %s", filepath.Base(path), msg)
		}
		return nil
	default:
		return fmt.Errorf("archivo no soportado para extraccion: %s", path)
	}
}

func isArchiveFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".zip" || ext == ".rar"
}

func atoi(s string) int {

	n := 0

	for _, c := range s {

		n = n*10 + int(c-'0')
	}

	return n
}

type seriesPage struct {
	URL  string
	Body string
}

func collectSeriesPages(client *http.Client, initialURL string, initialBody string) ([]seriesPage, error) {
	initialURL = strings.TrimSpace(initialURL)
	if initialURL == "" {
		return nil, fmt.Errorf("url inicial de serie vacia")
	}

	queue := []string{initialURL}
	queued := map[string]bool{initialURL: true}
	visited := make(map[string]bool)
	pages := make([]seriesPage, 0, 8)

	for len(queue) > 0 {
		currentURL := queue[0]
		queue = queue[1:]
		if visited[currentURL] {
			continue
		}
		visited[currentURL] = true

		var body string
		if currentURL == initialURL && strings.TrimSpace(initialBody) != "" {
			body = initialBody
		} else {
			fetched, err := fetchSeriesPageBody(client, currentURL, initialURL)
			if err != nil {
				fmt.Printf("[WARN] No se pudo leer página %s: %v\n", currentURL, err)
				continue
			}
			body = fetched
		}

		pages = append(pages, seriesPage{URL: currentURL, Body: body})

		nextLinks := extractSeriesPageLinks(body, currentURL)
		for _, nextURL := range nextLinks {
			if nextURL == "" || visited[nextURL] || queued[nextURL] {
				continue
			}
			queued[nextURL] = true
			queue = append(queue, nextURL)
		}
	}

	fmt.Printf("[DEBUG] Paginacion detectada: %d página(s) recorridas\n", len(pages))
	return pages, nil
}

func fetchSeriesPageBody(client *http.Client, pageURL string, referer string) (string, error) {
	resp, err := doGetWithRetry(client, pageURL, referer)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)

	if isVerificationPage(body) {
		flareBody, fsErr := solveWithFlareSolverr(client, pageURL)
		if fsErr != nil {
			return "", fmt.Errorf("anti-bot detectado en paginacion: %w", fsErr)
		}

		resp2, err := doGetWithRetry(client, pageURL, BaseURL+"/")
		if err == nil {
			defer resp2.Body.Close()
			bodyBytes2, readErr := io.ReadAll(resp2.Body)
			if readErr == nil {
				body = string(bodyBytes2)
			}
		}

		if isVerificationPage(body) && strings.TrimSpace(flareBody) != "" {
			body = flareBody
		}
	}

	if isIngresoGatePage(body) {
		if err := solveEntryGate(client); err != nil {
			return "", err
		}

		resp3, err := doGetWithRetry(client, pageURL, BaseURL+"/ingreso/")
		if err != nil {
			return "", err
		}
		defer resp3.Body.Close()

		bodyBytes3, err := io.ReadAll(resp3.Body)
		if err != nil {
			return "", err
		}
		body = string(bodyBytes3)
	}

	if isVerificationPage(body) {
		solved, newBody, err := solveCustomVerificationChallenge(client, body, pageURL)
		if err != nil {
			fmt.Printf("[DEBUG] Emulación challenge en paginacion falló: %v\n", err)
		}
		if solved {
			body = newBody
		}
	}

	return body, nil
}

func extractSeriesPageLinks(body string, pageURL string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil
	}

	seen := make(map[string]bool)
	links := make([]string, 0, 8)
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}

		text := strings.ToLower(strings.TrimSpace(s.Text()))
		h := strings.ToLower(href)
		if !strings.Contains(text, "siguiente") && !strings.Contains(h, "&pn=") && !strings.Contains(h, "?pn=") {
			return
		}

		abs := resolveRelativeURL(pageURL, href)
		if abs == "" || seen[abs] {
			return
		}
		seen[abs] = true
		links = append(links, abs)
	})

	return links
}

func parseEpisodesFromPages(pages []seriesPage) []episode {
	result := make([]episode, 0)
	seen := make(map[string]bool)

	for _, page := range pages {
		parsed := parseEpisodeBlocks(page.Body, page.URL)
		for _, ep := range parsed {
			key := strings.TrimSpace(ep.DownloadURL)
			if key == "" {
				key = strings.TrimSpace(ep.PageURL) + "|" + strings.TrimSpace(ep.Title) + "|" + strings.TrimSpace(ep.Filename)
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, ep)
		}
	}

	fmt.Printf("[DEBUG] Episodios parseados por bloques DOM: %d\n", len(result))
	return result
}

func parseEpisodeBlocks(body string, pageURL string) []episode {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		fmt.Printf("[WARN] No se pudo parsear HTML de %s: %v\n", pageURL, err)
		return nil
	}

	episodes := make([]episode, 0)
	doc.Find("span#animetitu").Each(func(_ int, titleSel *goquery.Selection) {
		title := normalizeSpace(titleSel.Text())
		titleEpisode := extractEpisodeNumberFromText(title)

		tr := titleSel.ParentsFiltered("tr").First()
		if tr.Length() == 0 {
			return
		}

		filename := ""
		downloadURL := ""
		editURL := ""

		for row := tr.Next(); row.Length() > 0; row = row.Next() {
			if row.Find("span#animetitu").Length() > 0 {
				break
			}

			if filename == "" {
				if occ := row.Find("div.ocultandoletras").First(); occ.Length() > 0 {
					if v, ok := occ.Attr("title"); ok && strings.TrimSpace(v) != "" {
						filename = normalizeSpace(v)
					} else {
						filename = normalizeSpace(occ.Text())
					}
				}
			}

			row.Find("input").Each(func(_ int, inputSel *goquery.Selection) {
				onclick, ok := inputSel.Attr("onClick")
				if !ok {
					onclick, ok = inputSel.Attr("onclick")
				}
				if !ok || strings.TrimSpace(onclick) == "" {
					return
				}

				if downloadURL == "" {
					if rel := extractOnClickArg(onclick, "descarga"); rel != "" {
						downloadURL = resolveRelativeURL(pageURL, rel)
					}
				}
				if editURL == "" {
					if rel := extractOnClickArg(onclick, "otro"); rel != "" {
						editURL = resolveRelativeURL(pageURL, rel)
					}
				}
			})

			if filename != "" && downloadURL != "" && editURL != "" {
				break
			}
		}

		if downloadURL == "" {
			fmt.Printf("[WARN] Episodio sin link descarga (bloque omitido): title=%q page=%s\n", title, pageURL)
			return
		}
		if filename == "" {
			fmt.Printf("[WARN] Episodio sin filename en bloque: title=%q page=%s\n", title, pageURL)
		}

		filenameEpisode := extractEpisodeNumberFromText(filename)
		if titleEpisode > 0 && filenameEpisode > 0 && titleEpisode != filenameEpisode {
			fmt.Printf("[WARN] mismatch: title episode=%d filename episode=%d title=%q filename=%q\n", titleEpisode, filenameEpisode, title, filename)
		}

		token := extractTokenFromDownloadURL(downloadURL)
		episodes = append(episodes, episode{
			Title:         title,
			EpisodeNumber: titleEpisode,
			Filename:      filename,
			DownloadURL:   downloadURL,
			EditURL:       editURL,
			PageURL:       pageURL,
			Token:         token,
		})

		fmt.Printf("[DEBUG] Episode block: title=%q file=%q download=%s\n", title, filename, downloadURL)
	})

	return episodes
}

func normalizeSpace(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func extractOnClickArg(onclick string, fn string) string {
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(fn) + `\(\s*['\"]([^'\"]+)['\"]`)
	m := re.FindStringSubmatch(onclick)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func resolveRelativeURL(baseURL string, rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return ""
	}

	b, err := url.Parse(baseURL)
	if err != nil {
		return rel
	}
	r, err := url.Parse(rel)
	if err != nil {
		return rel
	}
	return b.ResolveReference(r).String()
}

func extractEpisodeNumberFromText(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	if m := regexp.MustCompile(`(?i)s\s*([0-9]{1,2})\s*e\s*([0-9]{1,3})`).FindStringSubmatch(text); len(m) >= 3 {
		n, err := strconv.Atoi(m[2])
		if err == nil {
			return n
		}
	}

	if m := regexp.MustCompile(`(?i)sub\D*([0-9]{1,3})`).FindStringSubmatch(text); len(m) >= 2 {
		n, err := strconv.Atoi(m[1])
		if err == nil {
			return n
		}
	}

	if m := regexp.MustCompile(`(?i)(?:episodio|ep|e)\D*([0-9]{1,3})`).FindStringSubmatch(text); len(m) >= 2 {
		n, err := strconv.Atoi(m[1])
		if err == nil {
			return n
		}
	}

	nums := regexp.MustCompile(`([0-9]{1,3})`).FindAllStringSubmatch(text, -1)
	for i := len(nums) - 1; i >= 0; i-- {
		n, err := strconv.Atoi(nums[i][1])
		if err == nil && n > 0 && n < 400 && n != 720 && n != 1080 && n != 2160 {
			return n
		}
	}

	return 0
}

func extractTokenFromDownloadURL(downloadURL string) string {
	u, err := url.Parse(strings.TrimSpace(downloadURL))
	if err != nil {
		return ""
	}
	if u.Path == "" {
		return ""
	}
	token := strings.TrimSpace(path.Base(u.Path))
	if token == "." || token == "/" {
		return ""
	}
	return token
}

func selectEpisodes(eps []episode) []episode {
	fmt.Printf("\n[+] Se encontraron %d episodio(s):\n\n", len(eps))
	for i, ep := range eps {
		display := strings.TrimSpace(ep.Title)
		if display == "" {
			display = strings.TrimSpace(ep.Filename)
		}
		if display == "" {
			display = strings.TrimSpace(ep.DownloadURL)
		}
		fmt.Printf("  [%2d] %s\n", i+1, display)
	}
	fmt.Print("\nElige cuáles descargar (ej: 1 3 5-7) o 'todo' para todos: ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)

	if line == "" || line == "todo" || line == "all" {
		return eps
	}

	selected := map[int]bool{}
	for _, part := range strings.Fields(line) {
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			a, err1 := strconv.Atoi(bounds[0])
			b, err2 := strconv.Atoi(bounds[1])
			if err1 == nil && err2 == nil {
				for n := a; n <= b && n <= len(eps); n++ {
					selected[n] = true
				}
			}
		} else {
			n, err := strconv.Atoi(part)
			if err == nil && n >= 1 && n <= len(eps) {
				selected[n] = true
			}
		}
	}

	var result []episode
	for i, ep := range eps {
		if selected[i+1] {
			result = append(result, ep)
		}
	}
	return result
}

var reAssTag = regexp.MustCompile(`\{[^}]*\}`)

func assTimeToSRT(t string) string {
	// ASS: H:MM:SS.cc  →  SRT: HH:MM:SS,mmm
	parts := strings.Split(t, ":")
	if len(parts) != 3 {
		return t
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	secParts := strings.Split(parts[2], ".")
	s, _ := strconv.Atoi(secParts[0])
	cs := 0
	if len(secParts) > 1 {
		cs, _ = strconv.Atoi(secParts[1])
	}
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, cs*10)
}

func stripAssTags(text string) string {
	text = reAssTag.ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, `\N`, "\n")
	text = strings.ReplaceAll(text, `\n`, "\n")
	text = strings.ReplaceAll(text, `\h`, " ")
	return strings.TrimSpace(text)
}

func convertAssToSRT(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	fieldMap := map[string]int{}
	inEvents := false
	var entries []string
	idx := 1

	for _, line := range lines {
		if strings.EqualFold(strings.TrimSpace(line), "[Events]") {
			inEvents = true
			continue
		}
		if !inEvents {
			continue
		}
		if strings.HasPrefix(line, "[") {
			break
		}
		if strings.HasPrefix(line, "Format:") {
			for i, f := range strings.Split(strings.TrimPrefix(line, "Format:"), ",") {
				fieldMap[strings.TrimSpace(f)] = i
			}
			continue
		}
		if !strings.HasPrefix(line, "Dialogue:") {
			continue
		}

		nFields := len(fieldMap)
		if nFields == 0 {
			nFields = 10
		}
		parts := strings.SplitN(strings.TrimPrefix(line, "Dialogue:"), ",", nFields)
		if len(parts) < nFields {
			continue
		}

		startIdx := fieldMap["Start"]
		endIdx := fieldMap["End"]
		textIdx := nFields - 1
		if i, ok := fieldMap["Text"]; ok {
			textIdx = i
		}

		start := assTimeToSRT(strings.TrimSpace(parts[startIdx]))
		end := assTimeToSRT(strings.TrimSpace(parts[endIdx]))
		text := stripAssTags(strings.TrimSpace(parts[textIdx]))

		if text == "" {
			continue
		}

		entries = append(entries, fmt.Sprintf("%d\n%s --> %s\n%s", idx, start, end, text))
		idx++
	}

	return os.WriteFile(dst, []byte(strings.Join(entries, "\n\n")+"\n"), 0644)
}

func convertDirToSRT(dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".ass" && ext != ".ssa" {
			return nil
		}
		srtPath := path[:len(path)-len(ext)] + ".srt"
		fmt.Printf("[+] Convirtiendo a SRT: %s\n", filepath.Base(path))
		if cerr := convertAssToSRT(path, srtPath); cerr != nil {
			fmt.Printf("[!] Error convirtiendo %s: %v\n", filepath.Base(path), cerr)
			return nil
		}
		os.Remove(path)
		return nil
	})
	if err != nil {
		fmt.Printf("[!] Error recorriendo directorio: %v\n", err)
	}
}

func scanSubtitleCounts(dir string) (int, int, int, error) {
	srtCount := 0
	assCount := 0
	ssaCount := 0

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		switch strings.ToLower(filepath.Ext(path)) {
		case ".srt":
			srtCount++
		case ".ass":
			assCount++
		case ".ssa":
			ssaCount++
		}

		return nil
	})

	return srtCount, assCount, ssaCount, err
}

func extractPendingZips(dir string) (int, error) {
	archivePaths := make([]string, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if isArchiveFile(path) {
			archivePaths = append(archivePaths, path)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	if len(archivePaths) == 0 {
		return 0, nil
	}

	sort.Strings(archivePaths)

	count := 0
	for _, archivePath := range archivePaths {
		if err := extractArchiveFile(archivePath, dir); err != nil {
			fmt.Printf("[DEBUG] No se pudo extraer archivo pendiente %s: %v\n", archivePath, err)
			continue
		}

		if err := os.Remove(archivePath); err != nil {
			fmt.Printf("[DEBUG] Archivo extraido pero no se pudo borrar %s: %v\n", archivePath, err)
		}

		fmt.Printf("[DEBUG] Archivo anidado extraido: %s\n", archivePath)
		count++
	}

	return count, nil
}

func extractAllPendingZips(dir string, maxRounds int) (int, error) {
	if maxRounds < 1 {
		maxRounds = 1
	}

	total := 0
	for round := 1; round <= maxRounds; round++ {
		count, err := extractPendingZips(dir)
		if err != nil {
			return total, err
		}
		if count == 0 {
			return total, nil
		}
		total += count
	}

	return total, fmt.Errorf("se alcanzo maximo de rondas de extraccion (%d)", maxRounds)
}

func askYesNo(prompt string, defaultNo bool) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	line, _ := reader.ReadString('\n')
	v := strings.ToLower(strings.TrimSpace(line))
	if v == "" {
		return !defaultNo
	}
	return v == "s" || v == "si" || v == "y" || v == "yes"
}

type browserItem struct {
	name  string
	path  string
	isDir bool
}

func chooseDirectoryInteractive(start string) (string, error) {
	return browseMediaPathInteractive(start, false)
}

func chooseVideoInteractive(start string) (string, error) {
	return browseMediaPathInteractive(start, true)
}

func browseMediaPathInteractive(start string, allowVideoSelection bool) (string, error) {
	root := strings.TrimSpace(start)
	if root == "" {
		root = mediaRootDefault
	}

	if st, err := os.Stat(root); err != nil || !st.IsDir() {
		root = "."
	}

	current, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		entries, err := os.ReadDir(current)
		if err != nil {
			return "", err
		}

		dirs := make([]browserItem, 0)
		videos := make([]browserItem, 0)
		for _, e := range entries {
			name := e.Name()
			fullPath := filepath.Join(current, name)

			if e.IsDir() {
				if strings.EqualFold(name, ".trickplay") {
					continue
				}
				dirs = append(dirs, browserItem{name: name, path: fullPath, isDir: true})
				continue
			}

			if isVideoFile(name) {
				videos = append(videos, browserItem{name: name, path: fullPath, isDir: false})
			}
		}

		sort.Slice(dirs, func(i, j int) bool {
			return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name)
		})
		sort.Slice(videos, func(i, j int) bool {
			return strings.ToLower(videos[i].name) < strings.ToLower(videos[j].name)
		})

		items := make([]browserItem, 0, len(dirs)+len(videos))
		items = append(items, dirs...)
		items = append(items, videos...)

		fmt.Printf("\n[+] Carpeta actual: %s\n", current)
		if len(items) == 0 {
			fmt.Println("    (sin subcarpetas ni videos)")
		} else {
			for i, item := range items {
				if item.isDir {
					fmt.Printf("  [%d] %s/\n", i+1, item.name)
				} else {
					fmt.Printf("  [%d] %s\n", i+1, item.name)
				}
			}
		}

		if allowVideoSelection {
			fmt.Print("Elige numero, '..' subir, '.' seleccionar carpeta, ruta, o 'q' cancelar: ")
		} else {
			fmt.Print("Elige numero, '..' subir, '.' seleccionar carpeta, ruta, o 'q' cancelar: ")
		}

		line, _ := reader.ReadString('\n')
		in := strings.TrimSpace(line)

		switch in {
		case ".":
			return current, nil
		case "..":
			current = filepath.Dir(current)
			continue
		case "q", "Q":
			return "", fmt.Errorf("cancelado por usuario")
		}

		if n, err := strconv.Atoi(in); err == nil {
			if n < 1 || n > len(items) {
				fmt.Println("[!] Opcion fuera de rango")
				continue
			}

			chosen := items[n-1]
			if chosen.isDir {
				current = chosen.path
				continue
			}

			if !allowVideoSelection {
				fmt.Println("[!] Selecciona una carpeta, no un video")
				continue
			}

			return chosen.path, nil
		}

		if in != "" {
			cand := in
			if !filepath.IsAbs(cand) {
				cand = filepath.Join(current, cand)
			}

			st, err := os.Stat(cand)
			if err != nil {
				fmt.Println("[!] Ruta invalida")
				continue
			}

			if st.IsDir() {
				if strings.EqualFold(filepath.Base(cand), ".trickplay") {
					fmt.Println("[!] Carpeta ignorada por filtro: .trickplay")
					continue
				}
				current = cand
				continue
			}

			if !allowVideoSelection {
				fmt.Println("[!] Ruta invalida (se esperaba carpeta)")
				continue
			}

			if !isVideoFile(cand) {
				fmt.Println("[!] El archivo no es un video soportado (.mkv/.mp4/.avi)")
				continue
			}

			return cand, nil
		}
	}
}

func moveAndRenameSubtitles(subsDir string, videosBaseDir string) (int, int, error) {
	subFiles, err := listSubtitleFiles(subsDir)
	if err != nil {
		return 0, 0, err
	}

	if len(subFiles) == 0 {
		return 0, 0, fmt.Errorf("no se encontraron subtitulos en %s", subsDir)
	}

	sort.Strings(subFiles)

	moved := 0
	skipped := 0
	for _, sub := range subFiles {
		season, ep, ok := extractEpisodeRef(filepath.Base(sub))
		if !ok {
			fmt.Printf("[!] No pude detectar episodio para: %s\n", filepath.Base(sub))
			skipped++
			continue
		}

		video, err := findVideoByEpisode(videosBaseDir, season, ep)
		if err != nil {
			if season >= 0 {
				fmt.Printf("[!] No video encontrado para S%02dE%02d (%s)\n", season, ep, filepath.Base(sub))
			} else {
				fmt.Printf("[!] No video encontrado para episodio %02d (%s)\n", ep, filepath.Base(sub))
			}
			skipped++
			continue
		}

		if err := placeSubtitleForVideo(sub, video, "spa", false, false); err != nil {
			fmt.Printf("[!] Error moviendo/procesando %s: %v\n", filepath.Base(sub), err)
			skipped++
			continue
		}

		fmt.Printf("[+] Procesado: %s\n", filepath.Base(sub))
		moved++
	}

	return moved, skipped, nil

}

func listSubtitleFiles(subsDir string) ([]string, error) {
	srtFiles := make([]string, 0)
	err := filepath.Walk(subsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if isSubtitleFile(path) {
			srtFiles = append(srtFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(srtFiles)
	return srtFiles, nil
}

func hasMultipleLanguageHints(srtFiles []string) bool {
	seen := map[string]bool{}
	for _, f := range srtFiles {
		lang := inferLanguageHint(filepath.Base(f))
		if lang == "" {
			continue
		}
		seen[lang] = true
		if len(seen) > 1 {
			return true
		}
	}
	return false
}

func inferLanguageHint(name string) string {
	l := strings.ToLower(name)
	hints := []struct {
		needle string
		tag    string
	}{
		{needle: "español", tag: "spa"},
		{needle: "espanol", tag: "spa"},
		{needle: "spanish", tag: "spa"},
		{needle: "lat", tag: "spa"},
		{needle: "english", tag: "eng"},
		{needle: "inglés", tag: "eng"},
		{needle: "ingles", tag: "eng"},
		{needle: "deutsch", tag: "de"},
		{needle: "german", tag: "de"},
		{needle: "forced", tag: "forced"},
		{needle: "sdh", tag: "sdh"},
	}

	for _, h := range hints {
		if strings.Contains(l, h.needle) {
			return h.tag
		}
	}

	return ""
}

func sanitizeLanguageTag(tag string) string {
	tag = strings.ToLower(strings.TrimSpace(tag))
	if tag == "" {
		return ""
	}
	re := regexp.MustCompile(`[^a-z0-9._-]`)
	return re.ReplaceAllString(tag, "")
}

func isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".mkv" || ext == ".mp4" || ext == ".avi"
}

func isSubtitleFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".srt" || ext == ".ass" || ext == ".ssa"
}

func buildSubtitleTargets(videoPath string, langTag string, sourcePath string) (string, string, bool, error) {
	ext := strings.ToLower(filepath.Ext(sourcePath))
	base := strings.TrimSuffix(videoPath, filepath.Ext(videoPath))

	switch ext {
	case ".srt":
		return base + "." + langTag + ".srt", "", false, nil
	case ".ass", ".ssa":
		originalTarget := base + "." + langTag + ext
		srtTarget := base + "." + langTag + ".srt"
		return originalTarget, srtTarget, true, nil
	default:
		return "", "", false, fmt.Errorf("extension de subtitulo no soportada: %s", ext)
	}
}

func placeSubtitleForVideo(sourcePath string, videoPath string, langTag string, overwrite bool, askOverwrite bool) error {
	originalTarget, srtTarget, needsSRT, err := buildSubtitleTargets(videoPath, langTag, sourcePath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(originalTarget); err == nil {
		if askOverwrite {
			if !askYesNo("[?] Ya existe target original. Sobrescribir? (s/N): ", true) {
				return fmt.Errorf("cancelado por usuario")
			}
			overwrite = true
		} else if !overwrite {
			return fmt.Errorf("ya existe target original: %s", originalTarget)
		}
	}

	if err := moveFile(sourcePath, originalTarget); err != nil {
		return fmt.Errorf("error moviendo original: %w", err)
	}

	if !needsSRT {
		return nil
	}

	if _, err := os.Stat(srtTarget); err == nil {
		if askOverwrite {
			if !askYesNo("[?] Ya existe target .srt derivado. Sobrescribir? (s/N): ", true) {
				return fmt.Errorf("cancelado por usuario")
			}
			overwrite = true
		} else if !overwrite {
			fmt.Printf("[!] Ya existe .srt derivado, se conserva sin sobrescribir: %s\n", srtTarget)
			return nil
		}
	}

	if err := generateSRTWithFFmpeg(originalTarget, srtTarget); err != nil {
		return err
	}

	return nil
}

func generateSRTWithFFmpeg(inputPath string, outputPath string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg no disponible en PATH")
	}

	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", inputPath, outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("ffmpeg fallo (%s -> %s): %s", filepath.Base(inputPath), filepath.Base(outputPath), msg)
	}

	fmt.Printf("[+] SRT derivado con ffmpeg: %s\n", filepath.Base(outputPath))
	return nil
}

func moveSubtitlesManual(subsDir string, videosBaseDir string) (int, int, error) {
	if videosBaseDir == "" {
		videosBaseDir = mediaRootDefault
	}

	reader := bufio.NewReader(os.Stdin)
	moved := 0

	for {
		srtFiles, err := listSubtitleFiles(subsDir)
		if err != nil {
			return moved, 0, err
		}

		if len(srtFiles) == 0 {
			fmt.Println("[+] No quedan subtitulos por mover.")
			return moved, 0, nil
		}

		fmt.Println("\n[+] Subtitulos listos para mover:")
		for i, s := range srtFiles {
			fmt.Printf("  [%d] %s\n", i+1, filepath.Base(s))
		}
		fmt.Println("  [q] salir")
		fmt.Print("Elige subtitle a procesar: ")

		line, _ := reader.ReadString('\n')
		in := strings.TrimSpace(line)
		if strings.EqualFold(in, "q") {
			return moved, len(srtFiles), nil
		}

		idx, err := strconv.Atoi(in)
		if err != nil || idx < 1 || idx > len(srtFiles) {
			fmt.Println("[!] Opcion invalida")
			continue
		}

		subtitle := srtFiles[idx-1]

		fmt.Println("[+] Selecciona el video destino...")
		video, err := chooseVideoInteractive(videosBaseDir)
		if err != nil {
			fmt.Printf("[!] Seleccion de video cancelada: %v\n", err)
			continue
		}

		defaultTag := inferLanguageHint(filepath.Base(subtitle))
		if defaultTag == "" {
			defaultTag = "spa"
		}

		fmt.Printf("Tag de idioma (ej: es, spa, eng, de, forced, sdh) [default=%s]: ", defaultTag)
		tagLine, _ := reader.ReadString('\n')
		tag := sanitizeLanguageTag(tagLine)
		if tag == "" {
			tag = defaultTag
		}

		originalTarget, srtTarget, needsSRT, tgtErr := buildSubtitleTargets(video, tag, subtitle)
		if tgtErr != nil {
			fmt.Printf("[!] No se pudo preparar destino: %v\n", tgtErr)
			continue
		}

		fmt.Println("\n[+] Creare:")
		fmt.Println(filepath.Base(originalTarget))
		if needsSRT {
			fmt.Println(filepath.Base(srtTarget))
		}

		if !askYesNo("Confirmar? (S/n): ", false) {
			fmt.Println("[!] Operacion cancelada")
			continue
		}

		if err := placeSubtitleForVideo(subtitle, video, tag, false, true); err != nil {
			fmt.Printf("[!] Error procesando %s: %v\n", subtitle, err)
			continue
		}

		fmt.Printf("[+] Procesado: %s\n", filepath.Base(subtitle))
		moved++
	}
}

func samePath(a string, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return filepath.Clean(aa) == filepath.Clean(bb)
}

func cleanupSubsWorkspace(root string) error {
	cleanRoot := filepath.Clean(root)
	if cleanRoot != "/data/media/subs" {
		return fmt.Errorf("ruta no permitida para limpieza: %s", cleanRoot)
	}

	entries, err := os.ReadDir(cleanRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		child := filepath.Join(cleanRoot, entry.Name())
		if err := os.RemoveAll(child); err != nil {
			return err
		}
	}

	return nil
}

func extractEpisodeRef(name string) (int, int, bool) {
	reSE := regexp.MustCompile(`(?i)s\s*([0-9]{1,2})\s*e\s*([0-9]{1,3})`)
	if m := reSE.FindStringSubmatch(name); len(m) >= 3 {
		s, errS := strconv.Atoi(m[1])
		e, errE := strconv.Atoi(m[2])
		if errS == nil && errE == nil && e > 0 {
			return s, e, true
		}
	}

	rePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)sub\D*([0-9]{1,3})`),
		regexp.MustCompile(`(?i)(?:episodio|ep|e)\D*([0-9]{1,3})`),
	}

	for _, re := range rePatterns {
		m := re.FindStringSubmatch(name)
		if len(m) >= 2 {
			n, err := strconv.Atoi(m[1])
			if err == nil && n > 0 {
				return -1, n, true
			}
		}
	}

	allNums := regexp.MustCompile(`([0-9]{1,3})`).FindAllStringSubmatch(name, -1)
	for i := len(allNums) - 1; i >= 0; i-- {
		n, err := strconv.Atoi(allNums[i][1])
		if err == nil && n > 0 && n < 400 {
			// Avoid common resolution/year noise when guessing from plain numbers.
			if n == 720 || n == 1080 || n == 2160 {
				continue
			}
			return -1, n, true
		}
	}

	return -1, 0, false
}

func findVideoByEpisode(baseDir string, season int, ep int) (string, error) {
	ep2 := fmt.Sprintf("e%02d", ep)
	ep3 := fmt.Sprintf("e%03d", ep)
	se := ""
	if season >= 0 {
		se = fmt.Sprintf("s%02de%02d", season, ep)
	}

	var matches []string
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".mkv" && ext != ".mp4" && ext != ".avi" {
			return nil
		}
		name := strings.ToLower(filepath.Base(path))
		if se != "" && strings.Contains(name, se) {
			matches = append(matches, path)
			return nil
		}
		if strings.Contains(name, ep2) || strings.Contains(name, ep3) {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("not found")
	}

	sort.Strings(matches)
	return matches[0], nil
}

func moveFile(src string, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	return os.Remove(src)
}

func runPostScript(script string, destination string, series string) error {
	cmd := exec.Command("/bin/sh", "-lc", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"NENSAYSUBS_DESTINATION="+destination,
		"NENSAYSUBS_SERIES="+series,
	)
	return cmd.Run()
}

func newHTTPClient(jar http.CookieJar) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   20 * time.Second,
			KeepAlive: 20 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 25 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		// Middleboxes over VPN often break h2 framing; force HTTP/1.1.
		ForceAttemptHTTP2: false,
		TLSNextProto:      map[string]func(string, *tls.Conn) http.RoundTripper{},
	}

	return &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   90 * time.Second,
	}
}

type flareSolverrRequest struct {
	Cmd        string `json:"cmd"`
	URL        string `json:"url"`
	MaxTimeout int    `json:"maxTimeout"`
	Session    string `json:"session,omitempty"`
}

type flareSolverrCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

type flareSolverrResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Session  string `json:"session"`
	Solution struct {
		URL       string               `json:"url"`
		Status    int                  `json:"status"`
		UserAgent string               `json:"userAgent"`
		Response  string               `json:"response"`
		Cookies   []flareSolverrCookie `json:"cookies"`
	} `json:"solution"`
}

func isVerificationPage(body string) bool {
	b := strings.ToLower(body)
	return strings.Contains(b, "espere mientras se verifica su solicitud") ||
		strings.Contains(b, "window.location.reload") && strings.Contains(b, "z0f76a1d14fd21a8fb5fd0d03e0fdc3d3cedae52f")
}

func isIngresoGatePage(body string) bool {
	b := strings.ToLower(body)
	return strings.Contains(b, "entrada a nensaysubs") &&
		strings.Contains(b, "ingresa el resultado") &&
		strings.Contains(b, "name=\"valor\"")
}

type challengeSubmission struct {
	Action string
	Method string
	Fields map[string]string
	Beacon string
}

func solveCustomVerificationChallenge(client *http.Client, body string, pageURL string) (bool, string, error) {
	if !isVerificationPage(body) {
		return false, body, nil
	}

	submission, err := extractChallengeSubmission(body)
	if err != nil {
		return false, body, err
	}

	if submission == nil {
		return false, body, fmt.Errorf("challenge sin payload util")
	}

	hasForm := submission.Action != "" && len(submission.Fields) > 0
	hasBeacon := strings.TrimSpace(submission.Beacon) != ""
	if !hasForm && !hasBeacon {
		return false, body, fmt.Errorf("challenge sin payload util")
	}

	if hasBeacon {
		beaconURL := resolveChallengeAction(pageURL, submission.Beacon)
		bResp, bErr := doGetWithRetry(client, beaconURL, pageURL)
		if bErr != nil {
			return false, body, fmt.Errorf("fallo enviando beacon anti-bot: %w", bErr)
		}
		io.Copy(io.Discard, bResp.Body)
		bResp.Body.Close()
	}

	actionURL := pageURL
	if hasForm {
		actionURL = resolveChallengeAction(pageURL, submission.Action)
	}

	form := url.Values{}
	for k, v := range submission.Fields {
		form.Set(k, v)
	}

	if hasForm {
		method := strings.ToUpper(strings.TrimSpace(submission.Method))
		if method == "" {
			method = "GET"
		}

		if method == "POST" {
			resp, err := doPostFormWithRetry(client, actionURL, form, pageURL)
			if err != nil {
				return false, body, err
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		} else {
			u, err := url.Parse(actionURL)
			if err != nil {
				return false, body, err
			}
			q := u.Query()
			for k, vals := range form {
				for _, v := range vals {
					q.Add(k, v)
				}
			}
			u.RawQuery = q.Encode()
			resp, err := doGetWithRetry(client, u.String(), pageURL)
			if err != nil {
				return false, body, err
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}

	time.Sleep(1200 * time.Millisecond)

	var lastBody string
	var lastErr error
	for attempt := 1; attempt <= 4; attempt++ {
		resp, err := doGetWithRetry(client, pageURL, actionURL)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt) * 700 * time.Millisecond)
			continue
		}

		newBodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt) * 700 * time.Millisecond)
			continue
		}

		newBody := string(newBodyBytes)
		lastBody = newBody

		if resp.StatusCode >= 500 || isBadGatewayPage(newBody) {
			lastErr = fmt.Errorf("status temporal tras challenge: %s", resp.Status)
			fmt.Printf("[DEBUG] Challenge post-fetch %d/4 devolvió %s; reintentando...\n", attempt, resp.Status)
			time.Sleep(time.Duration(attempt) * 900 * time.Millisecond)
			continue
		}

		if isVerificationPage(newBody) {
			lastErr = fmt.Errorf("challenge sigue activo")
			time.Sleep(time.Duration(attempt) * 700 * time.Millisecond)
			continue
		}

		if !looksLikeSeriesPage(newBody) {
			lastErr = fmt.Errorf("html no parece página de serie")
			time.Sleep(time.Duration(attempt) * 700 * time.Millisecond)
			continue
		}

		return true, newBody, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no se pudo validar página final tras challenge")
	}

	if strings.TrimSpace(lastBody) != "" {
		return false, lastBody, lastErr
	}

	return false, body, lastErr
}

func resolveChallengeAction(pageURL string, action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		return pageURL
	}
	if strings.HasPrefix(action, "http://") || strings.HasPrefix(action, "https://") {
		return action
	}

	base, err := url.Parse(pageURL)
	if err != nil {
		return BaseURL + action
	}

	ref, err := url.Parse(action)
	if err != nil {
		return BaseURL + action
	}

	return base.ResolveReference(ref).String()
}

func extractChallengeSubmission(body string) (*challengeSubmission, error) {
	scripts := regexp.MustCompile(`(?is)<script[^>]*>(.*?)</script>`).FindAllStringSubmatch(body, -1)
	if len(scripts) == 0 {
		return nil, fmt.Errorf("challenge sin scripts")
	}

	var script string
	for _, m := range scripts {
		if len(m) < 2 {
			continue
		}
		candidate := m[1]
		cl := strings.ToLower(candidate)
		if strings.Contains(cl, "hbla3o0puowl") && strings.Contains(cl, "z0f76a1d14fd21a8fb5fd0d03e0fdc3d3cedae52f") {
			script = candidate
			break
		}
	}

	if strings.TrimSpace(script) == "" {
		return nil, fmt.Errorf("script de challenge no encontrado")
	}

	vm := goja.New()
	prelude := `
var __challengeSubmit = null;
var __challengeXHR = null;
function Elem(tag){
  this.tagName = String(tag || "").toLowerCase();
  this.children = [];
  this.style = "";
  this.id = "";
  this.name = "";
  this.value = "";
  this.type = "";
  this.method = "GET";
  this.action = "";
}
Elem.prototype.appendChild = function(c){ this.children.push(c); return c; };
Elem.prototype.submit = function(){
  var fields = {};
  for (var i = 0; i < this.children.length; i++) {
    var c = this.children[i];
    if (c && c.name) { fields[String(c.name)] = String(c.value || ""); }
  }
  __challengeSubmit = {
    action: String(this.action || ""),
    method: String(this.method || "GET"),
    fields: fields
  };
};

var __dom = { hbla3o0puowl: new Elem("div") };

var document = {
  createElement: function(tag){ return new Elem(tag); },
  getElementById: function(id){
    if (!__dom[id]) { __dom[id] = new Elem("div"); }
    return __dom[id];
  },
  addEventListener: function(evt, cb){ if (evt === "DOMContentLoaded" && typeof cb === "function") { cb(); } },
  attachEvent: function(evt, cb){ if (typeof cb === "function") { cb(); } }
};

function PluginArray(){}
function Plugin(){}
function MimeTypeArray(){}
function MimeType(){}

function FakePluginArray(){}
FakePluginArray.prototype = PluginArray.prototype;

function FakeMimeTypeArray(){}
FakeMimeTypeArray.prototype = MimeTypeArray.prototype;

var plugins = new FakePluginArray();
plugins.length = 1;
var p0 = Object.create(Plugin.prototype);
plugins[0] = p0;

var mimeTypes = new FakeMimeTypeArray();
mimeTypes.length = 1;
var m0 = Object.create(MimeType.prototype);
mimeTypes[0] = m0;

var navigator = {
  webdriver: false,
  userAgent: "Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/125.0",
  appVersion: "5.0",
  language: "es-MX",
  languages: ["es-MX", "es", "en-US"],
	plugins: plugins,
	mimeTypes: mimeTypes
};

function XMLHttpRequest(){
	this._method = "GET";
	this._url = "";
	this.open = function(method, url){
		this._method = String(method || "GET");
		this._url = String(url || "");
	};
	this.send = function(){
		__challengeXHR = {
			method: this._method,
			url: this._url
		};
	};
}

function URLSearchParams(){ this._ = []; }
URLSearchParams.prototype.append = function(k, v){ this._.push([String(k), String(v)]); };
URLSearchParams.prototype.toString = function(){
  var out = [];
  for (var i = 0; i < this._.length; i++) {
    out.push(encodeURIComponent(this._[i][0]) + "=" + encodeURIComponent(this._[i][1]));
  }
  return out.join("&");
};

var window = {
  document: document,
  navigator: navigator,
  outerHeight: 900,
  outerWidth: 1600,
  XMLHttpRequest: XMLHttpRequest,
  URLSearchParams: URLSearchParams,
  addEventListener: document.addEventListener
};

function setTimeout(cb){ if (typeof cb === "function") { cb(); } return 1; }
`

	if _, err := vm.RunString(prelude); err != nil {
		return nil, fmt.Errorf("prelude js inválido: %w", err)
	}

	if _, err := vm.RunString(script); err != nil {
		return nil, fmt.Errorf("ejecución challenge js falló: %w", err)
	}

	v := vm.Get("__challengeSubmit")
	fields := map[string]string{}
	action := ""
	method := ""
	if !goja.IsUndefined(v) && !goja.IsNull(v) {
		m, ok := v.Export().(map[string]interface{})
		if ok {
			action, _ = m["action"].(string)
			method, _ = m["method"].(string)
			if rawFields, ok := m["fields"].(map[string]interface{}); ok {
				for k, rv := range rawFields {
					fields[k] = fmt.Sprintf("%v", rv)
				}
			}
		}
	}

	beacon := ""
	xv := vm.Get("__challengeXHR")
	if !goja.IsUndefined(xv) && !goja.IsNull(xv) {
		xm, ok := xv.Export().(map[string]interface{})
		if ok {
			beacon, _ = xm["url"].(string)
		}
	}

	if strings.TrimSpace(action) == "" && strings.TrimSpace(beacon) == "" {
		return nil, fmt.Errorf("challenge no generó submit ni beacon")
	}

	return &challengeSubmission{Action: action, Method: method, Fields: fields, Beacon: beacon}, nil
}

func isBadGatewayPage(body string) bool {
	b := strings.ToLower(body)
	return strings.Contains(b, "502 bad gateway") ||
		strings.Contains(b, "openresty") && strings.Contains(b, "bad gateway")
}

func looksLikeSeriesPage(body string) bool {
	b := strings.ToLower(body)
	if strings.Contains(b, "descarga('../senos/") {
		return true
	}
	if strings.Contains(b, "name=\"code\"") || strings.Contains(b, "captcha") {
		return true
	}
	if isIngresoGatePage(body) {
		return true
	}
	return false
}

func solveWithFlareSolverr(client *http.Client, targetURL string) (string, error) {
	fsURL := getFlareSolverrURL()

	for attempt := 1; attempt <= 2; attempt++ {
		session, err := ensureFlareSolverrSession(client)
		if err != nil {
			return "", err
		}

		payload := flareSolverrRequest{
			Cmd:        "request.get",
			URL:        targetURL,
			MaxTimeout: flareSolverrTimeout,
			Session:    session,
		}

		fsResp, rawBody, err := callFlareSolverr(client, fsURL, payload)
		if err != nil {
			return "", err
		}

		if strings.ToLower(fsResp.Status) != "ok" {
			msg := fsResp.Message
			if msg == "" {
				msg = string(rawBody)
			}
			low := strings.ToLower(msg)
			if strings.Contains(low, "session") && strings.Contains(low, "not") && attempt < 2 {
				runtimeFlareSession = ""
				continue
			}
			return "", fmt.Errorf("flaresolverr status=%s: %s", fsResp.Status, msg)
		}

		if fsResp.Solution.UserAgent != "" {
			runtimeUserAgent = fsResp.Solution.UserAgent
		}

		targetParsed, err := url.Parse(targetURL)
		if err != nil {
			return "", err
		}

		set := make([]*http.Cookie, 0, len(fsResp.Solution.Cookies))
		for _, c := range fsResp.Solution.Cookies {
			if c.Name == "" {
				continue
			}
			path := c.Path
			if path == "" {
				path = "/"
			}
			set = append(set, &http.Cookie{
				Name:     c.Name,
				Value:    c.Value,
				Path:     path,
				Domain:   c.Domain,
				HttpOnly: c.HTTPOnly,
				Secure:   c.Secure,
			})
		}

		if len(set) > 0 {
			client.Jar.SetCookies(targetParsed, set)
		}

		fmt.Printf("[+] FlareSolverr OK. Session=%s Cookies importadas: %d\n", runtimeFlareSession, len(set))
		return fsResp.Solution.Response, nil
	}

	return "", fmt.Errorf("flaresolverr no pudo resolver request.get")
}

func getFlareSolverrURL() string {
	fsURL := strings.TrimSpace(os.Getenv("NENSAYSUBS_FLARESOLVERR_URL"))
	if fsURL == "" {
		fsURL = "http://flaresolverr:8191/v1"
	}
	return fsURL
}

func callFlareSolverr(client *http.Client, fsURL string, payload flareSolverrRequest) (flareSolverrResponse, []byte, error) {
	var zero flareSolverrResponse

	data, err := json.Marshal(payload)
	if err != nil {
		return zero, nil, err
	}

	req, err := http.NewRequest("POST", fsURL, bytes.NewReader(data))
	if err != nil {
		return zero, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", runtimeUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return zero, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, nil, err
	}

	var fsResp flareSolverrResponse
	if err := json.Unmarshal(body, &fsResp); err != nil {
		return zero, body, fmt.Errorf("respuesta de FlareSolverr inválida: %w", err)
	}

	return fsResp, body, nil
}

func ensureFlareSolverrSession(client *http.Client) (string, error) {
	if runtimeFlareSession != "" {
		return runtimeFlareSession, nil
	}

	fsURL := getFlareSolverrURL()
	payload := flareSolverrRequest{Cmd: "sessions.create", MaxTimeout: flareSolverrTimeout}
	fsResp, rawBody, err := callFlareSolverr(client, fsURL, payload)
	if err != nil {
		return "", err
	}

	if strings.ToLower(fsResp.Status) != "ok" {
		msg := fsResp.Message
		if msg == "" {
			msg = string(rawBody)
		}
		return "", fmt.Errorf("sessions.create fallo: %s", msg)
	}

	if strings.TrimSpace(fsResp.Session) == "" {
		return "", fmt.Errorf("sessions.create sin session id")
	}

	runtimeFlareSession = fsResp.Session
	fmt.Printf("[DEBUG] FlareSolverr session creada: %s\n", runtimeFlareSession)
	return runtimeFlareSession, nil
}

func destroyFlareSolverrSession(client *http.Client) error {
	if runtimeFlareSession == "" {
		return nil
	}

	fsURL := getFlareSolverrURL()
	payload := flareSolverrRequest{Cmd: "sessions.destroy", Session: runtimeFlareSession, MaxTimeout: flareSolverrTimeout}
	fsResp, rawBody, err := callFlareSolverr(client, fsURL, payload)
	if err != nil {
		return err
	}

	if strings.ToLower(fsResp.Status) != "ok" {
		msg := fsResp.Message
		if msg == "" {
			msg = string(rawBody)
		}
		return fmt.Errorf("sessions.destroy fallo: %s", msg)
	}

	fmt.Printf("[DEBUG] FlareSolverr session cerrada: %s\n", runtimeFlareSession)
	runtimeFlareSession = ""
	return nil
}

func doGetWithRetry(client *http.Client, rawURL string, referer string) (*http.Response, error) {
	urls := []string{rawURL}
	if strings.HasPrefix(rawURL, "https://") {
		urls = append(urls, "http://"+strings.TrimPrefix(rawURL, "https://"))
	}

	var lastErr error
	for _, u := range urls {
		for attempt := 1; attempt <= maxHTTPRetries; attempt++ {
			req, err := http.NewRequest("GET", u, nil)
			if err != nil {
				return nil, err
			}
			setCommonHeaders(req, referer)

			resp, err := client.Do(req)
			if err == nil {
				return resp, nil
			}

			lastErr = err
			if !isRetryableNetErr(err) {
				break
			}

			fmt.Printf("[DEBUG] GET retry %d/%d failed for %s: %v\n", attempt, maxHTTPRetries, u, err)
			time.Sleep(time.Duration(attempt) * 400 * time.Millisecond)
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("request failed")
	}

	return nil, lastErr
}

func doPostFormWithRetry(client *http.Client, rawURL string, form url.Values, referer string) (*http.Response, error) {
	body := form.Encode()

	var lastErr error
	for attempt := 1; attempt <= maxHTTPRetries; attempt++ {
		req, err := http.NewRequest("POST", rawURL, strings.NewReader(body))
		if err != nil {
			return nil, err
		}

		setCommonHeaders(req, referer)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !isRetryableNetErr(err) {
			break
		}

		fmt.Printf("[DEBUG] POST retry %d/%d failed for %s: %v\n", attempt, maxHTTPRetries, rawURL, err)
		time.Sleep(time.Duration(attempt) * 400 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("post request failed")
	}

	return nil, lastErr
}

func setCommonHeaders(req *http.Request, referer string) {
	req.Header.Set("User-Agent", runtimeUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "es-MX,es;q=0.9,en-US;q=0.7,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
}

func isRetryableNetErr(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, io.EOF) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "eof") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "tls") ||
		strings.Contains(msg, "temporary")
}

func check(err error) {

	if err != nil {

		panic(err)
	}
}
