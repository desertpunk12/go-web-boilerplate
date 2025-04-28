tg:
	templ generate --watch --proxy="http://localhost:3000" --cmd="go run ./cmd/hrapp/main.go"

tcss:
	tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css -v --watch --minify
	
	
s:
	make -j2 tg tcss
