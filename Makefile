web:
	cd backend && go run ./cmd/web

web-dev:
	cd backend && nodemon -e go,yaml,xml --signal SIGTERM --exec 'go' run ./cmd/web

ui:
	yarn --cwd ./frontend dev --dotenv ./env/local.env

deploy-ui:
	cd frontend && yarn build --dotenv ./env/prod.env
	cd frontend && zip dist.zip -r .output && scp -r dist.zip user@server:~/my-app/frontend
	rm frontend/dist.zip
	ssh -tt user@server "cd ~/my-app/frontend && unzip -o dist.zip"
	ssh -tt user@server "sudo supervisorctl restart frontend"

deploy-web:
	cd backend && go build ./cmd/web
	scp ./backend/web ec2-user@13.51.174.79:/home/smart-note/backend/__web
	scp ./backend/confs/prod.yaml ec2-user@13.51.174.79:/home/smart-note/__web.yaml
	rm backend/web
	ssh -tt ec2-user@server "cd ~/my-app/backend && mv web _web && mv __web web"
	ssh -tt ec2-user@server "cd ~/my-app/backend && mv web.yaml _web.yaml && mv __web.yaml web.yaml"
	ssh -tt ec2-user@server "sudo supervisorctl restart backend"
