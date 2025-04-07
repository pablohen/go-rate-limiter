# Go Rate Limiter

### Running the server locally:

1. Clone the repository:

```zsh
git clone https://github.com/pablohen/go-rate-limiter.git
```

2. Create environment variables:

```zsh
cp cmd/server/.env.example cmd/server/.env
```

3. Create the containers:

```zsh
docker-compose up -d
```

4. Test IP-based limiting (should return 429 after 5 requests):

```zsh
for i in {1..6}; do curl http://localhost:8080; done
```

5. Test token-based limiting (should return 429 after 10 requests):

```zsh
for i in {1..11}; do curl -H "API_KEY: test-token" http://localhost:8080; done
```
