web:
	cd backend && go run ./cmd/web

web-dev:
	cd backend && nodemon -e go,yaml,xml --signal SIGTERM --exec 'go' run ./cmd/web

ui:
	yarn --cwd ./frontend dev --dotenv ./env/local.env

dep-ui:
	cd frontend && yarn generate --dotenv ./env/prod.env
	scp -r frontend/.output/public/* deploy@152.42.167.146:/var/www/smart-note

dep-web:
	cd backend && go build ./cmd/web
	scp ./backend/web deploy@152.42.167.146:/opt/smart-note/backend/__web
	scp ./backend/confs/prod.yaml deploy@152.42.167.146:/opt/smart-note/backend/__prod.yaml
	rm backend/web
	ssh -tt deploy@152.42.167.146 "cd /opt/smart-note/backend && mv web _web && mv __web web"
	ssh -tt deploy@152.42.167.146 "cd /opt/smart-note/backend && mv prod.yaml _prod.yaml && mv __prod.yaml prod.yaml"
	ssh -tt deploy@152.42.167.146 "sudo systemctl restart smart-note"
