#!/bin/bash

set -e

SUBS_DIR="/data/media/subs"
MEDIA_ROOT="/data/media"
SYNC_ONLY=false

usage() {
	echo "Uso:"
	echo "  $0 <anime_slug>"
	echo "  $0 --sync-only"
	echo ""
	echo "Opciones:"
	echo "  -s, --sync-only   No descarga; abre directamente modo de sincronizacion"
	echo "  -h, --help        Muestra esta ayuda"
}

ANIME=""
while [ $# -gt 0 ]; do
	case "$1" in
		-s|--sync-only)
			SYNC_ONLY=true
			shift
			;;
		-h|--help)
			usage
			exit 0
			;;
		-*)
			echo "[!] Flag desconocido: $1"
			usage
			exit 1
			;;
		*)
			if [ -z "$ANIME" ]; then
				ANIME="$1"
			else
				echo "[!] Argumento extra no esperado: $1"
				usage
				exit 1
			fi
			shift
			;;
	esac
done

if [ "$SYNC_ONLY" = false ] && [ -z "$ANIME" ]; then
	usage
	exit 1
fi

if [ "$SYNC_ONLY" = true ]; then
	echo
	echo "[+] Modo sync-only activado: se omite descarga."
else
	echo
	echo "[+] Descargando subtitulos..."
	docker exec -it nensaysubs /usr/local/bin/nensaysubs "$ANIME" "$SUBS_DIR"

	echo
	read -rp "[?] Sincronizar subtitulos? (s/N): " NEED_SYNC

	[[ ! "$NEED_SYNC" =~ ^[Ss]$ ]] && exit 0
fi

find_video_for_sub() {
	local sub_path="$1"
	local dir
	local sub_file
	local stem
	local stripped
	local tag

	dir="$(dirname "$sub_path")"
	sub_file="$(basename "$sub_path")"
	stem="${sub_file%.*}"
	stripped="$stem"
	tag="${stem##*.}"

	# Si termina en tag comun (spa, eng, sdh, forced, etc), intenta sin ese sufijo.
	if [[ "$tag" =~ ^[A-Za-z0-9_-]{2,12}$ ]] && [ "$tag" != "$stem" ]; then
		stripped="${stem%.*}"
	fi

	for base in "$stem" "$stripped"; do
		for ext in mkv mp4 avi; do
			if [ -f "$dir/$base.$ext" ]; then
				echo "$dir/$base.$ext"
				return 0
			fi
		done
	done

	return 1
}

sync_single_pair() {
	local video="$1"
	local sub="$2"
	local ext
	local base
	local output

	ext="${sub##*.}"
	base="${sub%.*}"
	output="${base}.synced.${ext}"

	echo >&2
	echo "[+] Sincronizando subtitulos..." >&2
	echo "    VIDEO : $video" >&2
	echo "    SUB   : $sub" >&2
	echo "    OUT   : $output" >&2
	echo >&2

	docker compose run --rm ffsubsync "$video" -i "$sub" -o "$output"

	echo >&2
	echo "[+] Sincronización completada:" >&2
	echo "    $output" >&2
}

select_file() {
	local START_DIR="$1"
	local MODE="$2"
	local CURRENT="$START_DIR"

	while true; do
		echo >&2
		echo "[+] Carpeta actual: $CURRENT" >&2

		mapfile -t DIRS < <(
			find "$CURRENT" -maxdepth 1 -type d \
				! -name "*.trickplay" \
				! -path "$CURRENT" \
				| sort
		)

		if [ "$MODE" = "video" ]; then
			mapfile -t FILES < <(
				find "$CURRENT" -maxdepth 1 -type f \
					\( -iname "*.mkv" -o -iname "*.mp4" -o -iname "*.avi" \) \
					| sort
			)
		else
			mapfile -t FILES < <(
				find "$CURRENT" -maxdepth 1 -type f \
					\( -iname "*.srt" -o -iname "*.ass" -o -iname "*.ssa" \) \
					| sort
			)
		fi

		local IDX=1

		for d in "${DIRS[@]}"; do
			echo "  [$IDX] $(basename "$d")/" >&2
			IDX=$((IDX+1))
		done

		for f in "${FILES[@]}"; do
			echo "  [$IDX] $(basename "$f")" >&2
			IDX=$((IDX+1))
		done

		echo >&2
		read -rp "Elige numero, '..' subir, '.' cancelar: " CHOICE

		if [ "$CHOICE" = "." ]; then
			return 1
		fi

		if [ "$CHOICE" = ".." ]; then
			CURRENT="$(dirname "$CURRENT")"
			continue
		fi

		if ! [[ "$CHOICE" =~ ^[0-9]+$ ]]; then
			continue
		fi

		local NUM="$CHOICE"

		if [ "$NUM" -le "${#DIRS[@]}" ]; then
			CURRENT="${DIRS[$((NUM-1))]}"
			continue
		fi

		local FILE_INDEX=$((NUM - ${#DIRS[@]} - 1))

		if [ "$FILE_INDEX" -ge 0 ] && [ "$FILE_INDEX" -lt "${#FILES[@]}" ]; then
			echo "${FILES[$FILE_INDEX]}"
			return 0
		fi
	done
}

select_folder() {
	local CURRENT="$1"

	while true; do
		echo >&2
		echo "[+] Carpeta actual: $CURRENT" >&2

		mapfile -t DIRS < <(
			find "$CURRENT" -maxdepth 1 -type d \
				! -name "*.trickplay" \
				! -path "$CURRENT" \
				| sort
		)

		local IDX=1
		for d in "${DIRS[@]}"; do
			echo "  [$IDX] $(basename "$d")/" >&2
			IDX=$((IDX+1))
		done

		echo >&2
		read -rp "Elige numero, '..' subir, '.' seleccionar carpeta, 'q' cancelar: " CHOICE

		if [ "$CHOICE" = "q" ]; then
			return 1
		fi

		if [ "$CHOICE" = "." ]; then
			echo "$CURRENT"
			return 0
		fi

		if [ "$CHOICE" = ".." ]; then
			CURRENT="$(dirname "$CURRENT")"
			continue
		fi

		if ! [[ "$CHOICE" =~ ^[0-9]+$ ]]; then
			continue
		fi

		local NUM="$CHOICE"
		if [ "$NUM" -ge 1 ] && [ "$NUM" -le "${#DIRS[@]}" ]; then
			CURRENT="${DIRS[$((NUM-1))]}"
		fi
	done
}

sync_batch_folder() {
	local folder="$1"
	local mode="$2"
	local filter
	local synced=0
	local skipped=0

	case "$mode" in
		srt)
			filter='*.srt'
			;;
		ass)
			filter='*.ass'
			;;
		ssa)
			filter='*.ssa'
			;;
		all)
			filter='*.srt|*.ass|*.ssa'
			;;
		*)
			echo "[!] Modo no valido: $mode"
			return 1
			;;
	esac

	if [ "$mode" = "all" ]; then
		mapfile -t SUBS < <(find "$folder" -maxdepth 1 -type f \( -iname "*.srt" -o -iname "*.ass" -o -iname "*.ssa" \) | sort)
	else
		mapfile -t SUBS < <(find "$folder" -maxdepth 1 -type f -iname "$filter" | sort)
	fi

	if [ "${#SUBS[@]}" -eq 0 ]; then
		echo "[!] No se encontraron subtitulos en $folder para el filtro '$mode'."
		return 1
	fi

	echo
	echo "[+] Encontrados ${#SUBS[@]} subtitulos en: $folder"

	for sub in "${SUBS[@]}"; do
		if [[ "$sub" == *.synced.* ]]; then
			echo "[=] Omitido (ya sincronizado): $(basename "$sub")"
			skipped=$((skipped+1))
			continue
		fi

		if ! video=$(find_video_for_sub "$sub"); then
			echo "[=] Omitido (sin video homonimo): $(basename "$sub")"
			skipped=$((skipped+1))
			continue
		fi

		sync_single_pair "$video" "$sub"
		synced=$((synced+1))
	done

	echo
	echo "[+] Sync batch completado. sincronizados=$synced omitidos=$skipped"
}

echo
echo "[?] Modo de sincronización:"
echo "  [1] Manual (elegir video y subtitle uno por uno)"
echo "  [2] Lote por carpeta (todos los subs con video homonimo)"
read -rp "Elige modo [1/2] (default 1): " SYNC_MODE

if [ "$SYNC_MODE" = "2" ]; then
	echo
	echo "==============================="
	echo " Selecciona CARPETA"
	echo "==============================="
	FOLDER=$(select_folder "$MEDIA_ROOT") || {
		echo
		echo "[+] Done."
		exit 0
	}

	echo
	echo "[?] Que subtitulos sincronizar en lote?"
	echo "  [1] Solo .srt"
	echo "  [2] Solo .ass"
	echo "  [3] Solo .ssa"
	echo "  [4] Todos (.srt, .ass, .ssa)"
	read -rp "Elige opcion [1-4] (default 4): " SUB_MODE

	case "$SUB_MODE" in
		1) MODE="srt" ;;
		2) MODE="ass" ;;
		3) MODE="ssa" ;;
		*) MODE="all" ;;
	esac

	sync_batch_folder "$FOLDER" "$MODE"
	echo
	echo "[+] Done."
	exit 0
fi

while true; do
	echo
	echo "==============================="
	echo " Selecciona VIDEO"
	echo "==============================="

	VIDEO=$(select_file "$MEDIA_ROOT" "video") || break

	[ ! -f "$VIDEO" ] && echo "[!] Video no existe" && continue

	echo
	echo "[+] Video seleccionado:"
	echo "    $VIDEO"

	echo
	echo "==============================="
	echo " Selecciona SUBTITLE"
	echo "==============================="

	SUB=$(select_file "$MEDIA_ROOT" "subtitle") || break

	[ ! -f "$SUB" ] && echo "[!] Subtitle no existe" && continue

	echo
	echo "[+] Subtitle seleccionado:"
	echo "    $SUB"

	sync_single_pair "$VIDEO" "$SUB"

	echo
	read -rp "[?] Sincronizar otro? (s/N): " AGAIN

	[[ ! "$AGAIN" =~ ^[Ss]$ ]] && break
done

echo
echo "[+] Done."
