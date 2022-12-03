build:
	go build -buildvcs=false

token:
	@echo 'export TOKEN=$$(token timc apple review secret)'

test:
	@echo 'curl -v --header "Authorization: Bearer $(token -r timc dev review secret)" http://127.0.0.1:8080/baas/c'
