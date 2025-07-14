build_and_push:
	docker build -t youcef123/cert-scanner:dev-0.1.3 .
	docker push  youcef123/cert-scanner:dev-0.1.3
deploy:
	kubectl apply -f deploy.yaml