docker inspect --format "{{.Name}}" $1 | sed 's/\///g'
