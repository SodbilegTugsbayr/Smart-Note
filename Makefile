web:
	cd backend && go run ./cmd/web

ui:
	yarn --cwd ./frontend dev --dotenv ./env/local.env --host

deploy-ui:
	cd frontend && yarn build --dotenv ./env/prod.env
	cd frontend && zip dist.zip -r .output && scp -r dist.zip user@server:~/my-app/frontend
	rm frontend/dist.zip
	ssh -tt user@server "cd ~/my-app/frontend && unzip -o dist.zip"
	ssh -tt user@server "sudo supervisorctl restart frontend"

deploy-web:
	cd backend && go build ./cmd/web
	scp ./backend/web user@server:~/my-app/backend/__web
	scp ./backend/confs/prod.yaml user@server:~/my-app/backend/__web.yaml
	rm backend/web
	ssh -tt user@server "cd ~/my-app/backend && mv web _web && mv __web web"
	ssh -tt user@server "cd ~/my-app/backend && mv web.yaml _web.yaml && mv __web.yaml web.yaml"
	ssh -tt user@server "sudo supervisorctl restart backend"

