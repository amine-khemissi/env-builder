#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
export SCRIPT_DIR
CONFIG="$SCRIPT_DIR/tools.yaml"
CLEAN=false

for arg in "$@"; do
    case "$arg" in
        --clean) CLEAN=true ;;
        *) echo "Unknown option: $arg"; echo "Usage: $0 [--clean]"; exit 1 ;;
    esac
done

if [[ ! -f "$CONFIG" ]]; then
    echo "Error: $CONFIG not found."
    exit 1
fi

if ! command -v yq &>/dev/null; then
    echo "Installing yq (YAML parser)..."
    sudo pacman -S --noconfirm --needed yq
fi

run_post_install() {
    local manager="$1"
    local count
    count=$(yq ".$manager | length" "$CONFIG")
    for ((i = 0; i < count; i++)); do
        local post_count
        post_count=$(yq ".$manager[$i].post_install // [] | length" "$CONFIG")
        if [[ "$post_count" -eq 0 ]]; then
            continue
        fi
        local pkg_name
        pkg_name=$(yq ".$manager[$i].name" "$CONFIG")
        for ((j = 0; j < post_count; j++)); do
            local cmd
            cmd=$(yq ".$manager[$i].post_install[$j]" "$CONFIG")
            echo "[$pkg_name] post-install: $cmd"
            eval "$cmd"
        done
    done
}

install_section() {
    local manager="$1"
    local packages=()

    while IFS= read -r pkg; do
        packages+=("$pkg")
    done < <(yq ".$manager[].name" "$CONFIG")

    if [[ ${#packages[@]} -eq 0 ]]; then
        return
    fi

    echo "[$manager] Installing: ${packages[*]}"
    if [[ "$manager" == "pacman" ]]; then
        sudo pacman -S --noconfirm --needed "${packages[@]}"
    else
        paru -S --noconfirm --needed "${packages[@]}"
    fi

    run_post_install "$manager"
}

install_custom_section() {
    local count
    count=$(yq ".custom | length" "$CONFIG")
    for ((i = 0; i < count; i++)); do
        local pkg_name
        pkg_name=$(yq ".custom[$i].name" "$CONFIG")
        local step_count
        step_count=$(yq ".custom[$i].install // [] | length" "$CONFIG")
        if [[ "$step_count" -eq 0 ]]; then
            continue
        fi
        echo "[custom] Installing: $pkg_name"
        for ((j = 0; j < step_count; j++)); do
            local cmd
            cmd=$(yq ".custom[$i].install[$j]" "$CONFIG")
            echo "[$pkg_name] $cmd"
            eval "$cmd"
        done
    done
}

clean_section() {
    local manager="$1"
    local packages=()

    while IFS= read -r pkg; do
        packages+=("$pkg")
    done < <(yq ".$manager[].name" "$CONFIG")

    if [[ ${#packages[@]} -eq 0 ]]; then
        return
    fi

    echo "[$manager] Removing: ${packages[*]}"
    sudo pacman -Rns --noconfirm "${packages[@]}" 2>/dev/null || true
}

clean_custom_section() {
    local count
    count=$(yq ".custom | length" "$CONFIG")
    for ((i = 0; i < count; i++)); do
        local pkg_name
        pkg_name=$(yq ".custom[$i].name" "$CONFIG")
        local step_count
        step_count=$(yq ".custom[$i].uninstall // [] | length" "$CONFIG")
        if [[ "$step_count" -eq 0 ]]; then
            continue
        fi
        echo "[custom] Removing: $pkg_name"
        for ((j = 0; j < step_count; j++)); do
            local cmd
            cmd=$(yq ".custom[$i].uninstall[$j]" "$CONFIG")
            echo "[$pkg_name] $cmd"
            eval "$cmd"
        done
    done
}

if [[ "$CLEAN" == true ]]; then
    clean_custom_section
    clean_section paru
    clean_section pacman
else
    install_section pacman
    install_section paru
    install_custom_section
fi

echo "Done."
