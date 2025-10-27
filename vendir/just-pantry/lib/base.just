import 'partial/shell.just'
import 'partial/settings.just'
import 'partial/vars.just'
import 'partial/utils.just'

# Lists all available/public recipes.
[group('global')]
@list: just-list

# Lets the user select the recipe to run. Uses fzf for fuzzy search of recipes.
[group('global')]
@choose: just-choose

# Hepler recipe to proxy a command through just.
[group('global')]
@x command *args:
    {{ command }} {{ args }}

# Opens just online documentation in your default browser
[group('global')]
sync *args='': (requires 'vendir')
    vendir sync {{ args }}
