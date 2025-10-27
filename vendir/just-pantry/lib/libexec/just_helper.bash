NO_FORMAT="\033[0m"
C_BLUE="\033[38;5;12m"
C_RED="\033[38;5;9m"
C_YELLOW="\033[38;5;11m"
C_LIME="\033[38;5;10m"
C_GREY="\033[38;5;8m"
C_DARKORANGE="\033[38;5;208m"
C_SILVER="\033[38;5;7m"
F_BOLD="\033[1m"

# Helper for calling recipes from within other recipes
function @recipe() {
  $JUST_EXECUTABLE -f "${JUST_JUSTFILE}" "${@}"
}

# Prints debug messages if DEBUG environment variable is set to true
function @debug() {
    if [ "${DEBUG}" == "true" ]; then
      echo -e "${F_BOLD}${C_BLUE}DEBUG:${NO_FORMAT}" "$@"
    fi
}

# Prints info messages
function @info() {
    @colorize "silver" "$@"
}

# Prints warning messages
function @warn() {
    echo -e "${F_BOLD}${C_DARKORANGE}WARN:${NO_FORMAT}" "$@"
}

# Prints error messages
function @error() {
    echo -e "${F_BOLD}${C_RED}ERROR:${NO_FORMAT}" "$@"
}

# Prints error messages and exits with 1
function @fatal() {
  @error "${@}"
  exit 1
}

# Prints colored messages
# Usage:
#   @colorize "green" "My message"
function @colorize() {
  local color="${1}"; shift
  local color_code

  case $color in

    blue)
      color_code=$C_BLUE
      ;;
    red)
      color_code=$C_RED
      ;;
    yellow)
      color_code=$C_YELLOW
      ;;
    green)
      color_code=$C_LIME
      ;;
    grey)
      color_code=$C_GREY
      ;;
    silver)
      color_code=$C_SILVER
      ;;
    orange)
      color_code=$C_DARKORANGE
      ;;
    *)
      color_code=$NO_FORMAT
      ;;
  esac

  echo -e "${color_code}${*}${NO_FORMAT}"
}

# Prints a horizontal line
function @hr() {
  local color="${1:-grey}"
  @colorize "${color}" "$(printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -)"
}

# Add a line break
function @break() {
  echo -e "\n"
}

# Loads a given .env file
function @dotenv() {
  local env_file=${1:-.env};
  set -a && source "$env_file" && set +a
}
