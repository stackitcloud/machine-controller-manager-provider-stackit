########################################################################################################################
# Common variables and flags
########################################################################################################################
# to be used instead of $(pwd), as it allows to overwrite it for testing
# Reexport for helpers

export JUST_JUSTFILE := justfile()

# Reexport for helpers

export JUST_JUSTFILE_DIRECTORY := justfile_directory()

# Reexport for helper

export JUST_EXECUTABLE := just_executable()

# Export a global debug flag to be used in recipes to print additional information

export DEBUG := env("DEBUG", "")

# Holds the repository root directory path

export PROJECT_ROOT := shell("git rev-parse --show-toplevel")

# Takes the name of the folder where the main justfile is located as project name

export PROJECT_NAME := env("PROJECT_NAME", shell('basename $1', PROJECT_ROOT))

##############
# Internal vars
##############
# converted linux architecture.
# see: https://just.systems/man/en/functions.html#system-information
# see: https://stackoverflow.com/questions/31851611/differences-between-arm64-and-aarch64/47274698#47274698

linux_arch := if arch() == "x86_64" { "amd64" } else if arch() == "aarch64" { "arm64" } else { arch() }
