#!/bin/bash
set -e

# This script creates hard links for your configuration files.
# It will ask for confirmation before linking each category of configs.

# Get the absolute path of the directory where the script is located
SOURCE_ROOT=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)

# Function to ask for user confirmation
confirm() {
    read -r -p "${1} [y/N] " response
    case "$response" in
        [yY][eE][sS]|[yY])
            true
            ;;
        *)
            false
            ;;
    esac
}

# Function to create hard links for all files in a source directory to a destination directory
create_hard_links_from_dir() {
    local source_dir="$1"
    local dest_dir="$2"

    if [ ! -d "$source_dir" ]; then
        echo "Warning: Source directory $source_dir not found."
        return
    fi

    mkdir -p "$dest_dir"
    echo "Linking files from $source_dir to $dest_dir..."
    # Find all files in source and link them to destination, preserving directory structure.
    find "$source_dir" -type f | while read -r file_path; do
        local relative_path="${file_path#$source_dir/}"
        local dest_file="$dest_dir/$relative_path"
        echo "  - Linking $relative_path"
        mkdir -p "$(dirname "$dest_file")"
        ln -f "$file_path" "$dest_file"
    done
    echo "Done linking files from $source_dir."
}

# --- Scripts ---
link_scripts() {
    local default_dest="$HOME/.scratch/scripts"
    read -p "Enter destination directory for scripts (default: $default_dest): " dest_dir
    dest_dir=${dest_dir:-$default_dest}
    create_hard_links_from_dir "$SOURCE_ROOT/scripts" "$dest_dir"
}

# --- Dotfiles in home directory ---
link_dotfiles() {
    local source_dir="$SOURCE_ROOT"
    local dest_dir="$HOME"
    echo "Linking dotfiles to $dest_dir..."

    # Files that are already dotfiles
    for file in .aliases .gitconfig .zshrc; do
        if [ -f "$source_dir/$file" ]; then
            echo "  - Linking $file"
            ln -f "$source_dir/$file" "$dest_dir/$file"
        else
            echo "  - Warning: $source_dir/$file not found."
        fi
    done

    # Files that need to be renamed to dotfiles
    if [ -f "$source_dir/tmux.conf" ]; then
        echo "  - Linking tmux.conf as .tmux.conf"
        ln -f "$source_dir/tmux.conf" "$dest_dir/.tmux.conf"
    else
        echo "  - Warning: $source_dir/tmux.conf not found."
    fi
    echo "Done linking dotfiles."
}

# --- Neovim (nvim) ---
link_nvim() {
    local default_dest="$HOME/.config/nvim"
    read -p "Enter destination directory for nvim config (default: $default_dest): " dest_dir
    dest_dir=${dest_dir:-$default_dest}
    local source_dir="$SOURCE_ROOT/nvim"

    if [ ! -d "$source_dir" ]; then
        echo "Warning: nvim source directory not found at $source_dir"
        return
    fi

    echo "Linking nvim config from $source_dir to $dest_dir..."
    # Find all files in source and link them to destination, preserving directory structure.
    find "$source_dir" -type f | while read -r file_path; do
        local relative_path="${file_path#$source_dir/}"
        local dest_file="$dest_dir/$relative_path"
        echo "  - Linking $relative_path"
        mkdir -p "$(dirname "$dest_file")"
        ln -f "$file_path" "$dest_file"
    done
    echo "Done linking nvim config."
}

# --- Tmuxinator ---
link_tmuxinator() {
    local default_dest="$HOME/.config/tmuxinator"
    read -p "Enter destination directory for tmuxinator configs (default: $default_dest): " dest_dir
    dest_dir=${dest_dir:-$default_dest}
    create_hard_links_from_dir "$SOURCE_ROOT/tmuxinator" "$dest_dir"
}

# --- Karabiner ---
link_karabiner() {
    local default_dest="$HOME/.config/karabiner"
    read -p "Enter destination directory for karabiner configs (default: $default_dest): " dest_dir
    dest_dir=${dest_dir:-$default_dest}
    create_hard_links_from_dir "$SOURCE_ROOT/karabiner" "$dest_dir"
}

# --- Main Execution Logic ---
echo "This script will help you set up your configuration files by creating hard links."
echo "You will be prompted to select a configuration category to link."
echo

OPTIONS=("Scripts" "Dotfiles" "Nvim" "Tmuxinator" "Karabiner" "All" "Quit")

while true; do
    echo "Select an option to link:"
    PS3='Please enter your choice: '
    select opt in "${OPTIONS[@]}"; do
        case $opt in
            "Scripts")
                link_scripts
                echo
                break
                ;;
            "Dotfiles")
                link_dotfiles
                echo
                break
                ;;
            "Nvim")
                link_nvim
                echo
                break
                ;;
            "Tmuxinator")
                link_tmuxinator
                echo
                break
                ;;
            "Karabiner")
                link_karabiner
                echo
                break
                ;;
            "All")
                link_scripts
                link_dotfiles
                link_nvim
                link_tmuxinator
                link_karabiner
                echo
                break
                ;;
            "Quit")
                echo "Exiting."
                exit 0
                ;;
            *)
                echo "Invalid option $REPLY. Please select a number from 1 to ${#OPTIONS[@]}."
                break
                ;;
        esac
    done
    echo "----------------------------------------"
done
