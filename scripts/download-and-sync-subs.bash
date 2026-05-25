#!/bin/bash

set -e

SUBS_DIR="/data/media/subs"
MEDIA_ROOT="/data/media"

if [ $# -lt 1 ]; then
	echo "Uso:"
	echo "  $0 <anime_slug>"
	exit 1
fi

ANIME="$1"

echo
echo "[+] Descargando subtitulos..."
docker exec -it nensaysubs /usr/local/bin/nensaysubs "$ANIME" "$SUBS_DIR"

echo
read -rp "[?] Sincronizar subtitulos? (s/N): " NEED_SYNC

[[ ! "$NEED_SYNC" =~ ^[Ss]$ ]] && exit 0

select_file() {
	local START_DIR="$1"
	local MODE="$2"
	local CURRENT="$START_DIR"

	while true; do
		echo
		echo "[+] Carpeta actual: $CURRENT"

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
			echo "  [$IDX] $(basename "$d")/"
			IDX=$((IDX+1))
		done

		for f in "${FILES[@]}"; do
			echo "  [$IDX] $(basename "$f")"
			IDX=$((IDX+1))
		done

		echo
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

	EXT="${SUB##*.}"
	BASE="${SUB%.*}"
	OUTPUT="${BASE}.synced.${EXT}"

	echo
	echo "[+] Sincronizando subtitulos..."
	echo "    VIDEO : $VIDEO"
	echo "    SUB   : $SUB"
	echo "    OUT   : $OUTPUT"
	echo

	docker compose run --rm ffsubsync "$VIDEO" -i "$SUB" -o "$OUTPUT"

	echo
	echo "[+] Sincronización completada:"
	echo "    $OUTPUT"

	echo
	read -rp "[?] Sincronizar otro? (s/N): " AGAIN

	[[ ! "$AGAIN" =~ ^[Ss]$ ]] && break
done

echo
echo "[+] Done."
