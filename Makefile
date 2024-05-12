all: templ tailwind build

.PHONY: templ
templ:
	@templ generate

.PHONY: tailwind
tailwind: assets/dist/styles.css

assets/dist/styles.css: tailwind.config.js assets/css/tailwind.css
	@npm run tailwind
	@touch assets/dist/styles.css

.PHONY: air
air: templ
	@go build -o ./tmp/main .

.PHONY: build
build:
	go build -o kasama-twitch-integrations .
