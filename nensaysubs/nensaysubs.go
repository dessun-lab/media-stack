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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/otiai10/gosseract/v2"
	"golang.org/x/image/draw"
)

const (
	BaseURL   = "https://nensaysubs.net"
	UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0"
)

const captchaPreviewWidth = 80
const maxHTTPRetries = 3

var runtimeUserAgent = UserAgent

type episode struct {
	name  string
	token string
}

func main() {
	args := os.Args[1:]

	series := ""
	if len(args) > 0 {
		series = args[0]
	} else {
		series = strings.TrimSpace(os.Getenv("NENSAYSUBS_SERIES"))
	}

	if series == "" {
		fmt.Fprintln(os.Stderr, "uso: nensaysubs <SerieSlug> [destino]")
		fmt.Fprintln(os.Stderr, "     nensaysubs Maoyuu_Maou_Yuusha /data/media/subs")
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

func runSeriesDownload(series string, destination string, interactive bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during download: %v", r)
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
		if err := solveWithFlareSolverr(client, seriesURL); err != nil {
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
	}

	episodes := parseEpisodes(body)
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
		fmt.Printf("\n[+] Procesando %d/%d: %s\n", idx+1, len(episodes), ep.name)

		token := ep.token
		episodeURL := BaseURL + "/senos/" + token
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

		convertDirToSRT(destination)

		downloadedCount++
		fmt.Println("[+] Episodio completado.")
	}

	if downloadedCount == 0 {
		return fmt.Errorf("no se pudo descargar ningún episodio")
	}

	fmt.Printf("[+] Listo. Descargados %d/%d episodios.\n", downloadedCount, len(episodes))

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

func atoi(s string) int {

	n := 0

	for _, c := range s {

		n = n*10 + int(c-'0')
	}

	return n
}

func parseEpisodes(body string) []episode {
	// Intenta capturar el nombre del episodio del HTML que rodea a descarga().
	reNamed := regexp.MustCompile(`(?s)>([^<\n]{2,80}?)</[a-z]+[^>]*>\s*[^<]*?descarga\('\.\./senos/([a-zA-Z0-9]+)'`)
	namedMatches := reNamed.FindAllStringSubmatch(body, -1)
	nameFor := map[string]string{}
	for _, m := range namedMatches {
		if len(m) >= 3 {
			name := strings.TrimSpace(m[1])
			token := m[2]
			if name != "" && nameFor[token] == "" {
				nameFor[token] = name
			}
		}
	}

	reToken := regexp.MustCompile(`descarga\('\.\./senos/([a-zA-Z0-9]+)'`)
	allMatches := reToken.FindAllStringSubmatch(body, -1)

	seen := map[string]bool{}
	var eps []episode
	for i, m := range allMatches {
		if len(m) < 2 {
			continue
		}
		token := m[1]
		if seen[token] {
			continue
		}
		seen[token] = true
		name := nameFor[token]
		if name == "" {
			name = fmt.Sprintf("Episodio %02d", i+1)
		}
		eps = append(eps, episode{name: name, token: token})
	}
	return eps
}

func selectEpisodes(eps []episode) []episode {
	fmt.Printf("\n[+] Se encontraron %d episodio(s):\n\n", len(eps))
	for i, ep := range eps {
		fmt.Printf("  [%2d] %s\n", i+1, ep.name)
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
	Solution struct {
		URL       string               `json:"url"`
		Status    int                  `json:"status"`
		UserAgent string               `json:"userAgent"`
		Cookies   []flareSolverrCookie `json:"cookies"`
	} `json:"solution"`
}

func isVerificationPage(body string) bool {
	b := strings.ToLower(body)
	return strings.Contains(b, "espere mientras se verifica su solicitud") ||
		strings.Contains(b, "window.location.reload") && strings.Contains(b, "z0f76a1d14fd21a8fb5fd0d03e0fdc3d3cedae52f")
}

func solveWithFlareSolverr(client *http.Client, targetURL string) error {
	fsURL := strings.TrimSpace(os.Getenv("NENSAYSUBS_FLARESOLVERR_URL"))
	if fsURL == "" {
		fsURL = "http://flaresolverr:8191/v1"
	}

	payload := flareSolverrRequest{
		Cmd:        "request.get",
		URL:        targetURL,
		MaxTimeout: 120000,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fsURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", runtimeUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var fsResp flareSolverrResponse
	if err := json.Unmarshal(body, &fsResp); err != nil {
		return fmt.Errorf("respuesta de FlareSolverr inválida: %w", err)
	}

	if strings.ToLower(fsResp.Status) != "ok" {
		if fsResp.Message == "" {
			fsResp.Message = string(body)
		}
		return fmt.Errorf("flaresolverr status=%s: %s", fsResp.Status, fsResp.Message)
	}

	if fsResp.Solution.UserAgent != "" {
		runtimeUserAgent = fsResp.Solution.UserAgent
	}

	targetParsed, err := url.Parse(targetURL)
	if err != nil {
		return err
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

	fmt.Printf("[+] FlareSolverr OK. Cookies importadas: %d\n", len(set))
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
