[private]
@just-list file=justfile():
    {{ just_executable() }} -f {{ file }} --list --list-submodules

[private]
@just-choose file=justfile(): (requires 'fzf')
    {{ just_executable() }} -f {{ file }} --choose

[no-cd]
[private]
@recipe name *args:
    {{ just_executable() }} -f {{ justfile() }} -d {{ justfile_directory() }} {{ name }} {{ args }}

[no-cd]
[private]
@recipe-file file name *args:
    {{ just_executable() }} -f {{ file }} -d {{ justfile_directory() }} {{ name }} {{ args }}

# Allows to disable requires checks entirely.

requires-enabled := env("JUST_PANTRY_REQUIRES_ENABLED", "true")

# Checks if the given programs are available in PATH
[private]
requires *programs:
    #!/usr/bin/env bash
    source "{{ source_directory() }}/../libexec/just_helper.bash"

    if [ "{{ requires-enabled }}" == "false" ]; then
      @debug "require checks are disabled!"
      exit 0
    fi

    if [ -z "{{ programs }}" ]; then
      @fatal "No arguments passed to requires recipe!"
    fi

    missing_programs=()
    for program in {{ programs }}; do
      command -v "${program}" >/dev/null 2>&1 || missing_programs+=("${program}")
    done

    if [ ${#missing_programs[@]} -gt 0 ]; then
      ## Print error messages for all missing programs
      echo -e "${C_BOLD}${C_RED}ERROR:${NO_FORMAT} Missing dependencies:"
      for e in "${missing_programs[@]}"; do
        echo -e "\t${e}"
      done

      @info "Install missing dependencies via hermit"
      @info "hermit install ${missing_programs[@]}"

      exit 1
    fi

# Checks if the given variables are set (not empty or undefined)
[private]
requires-vars *vars:
    #!/usr/bin/env bash
    source "{{ source_directory() }}/../libexec/just_helper.bash"
    missing_vars=()
    for var in {{ vars }}; do
      value="${!var}"
      if [ -z "$value" ]; then
        missing_vars+=("$var")
      fi
    done
    if [ ${#missing_vars[@]} -gt 0 ]; then
      echo -e "${C_BOLD}${C_RED}ERROR:${NO_FORMAT} Missing required variables: ${missing_vars[*]}"
      exit 1
    fi
