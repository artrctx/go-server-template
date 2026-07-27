# shuffle core

```bash
# sqlc dep
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
# migrate dep
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

![shuffle](./public/shuffle.gif)

---

```bash
# Add these variables to github env
# secrets
BETTER_AUTH_SECRET
DOCKER_TOKEN
DROPLET_PASSWORD
DATABASE_URL
DIRECT_DATABASE_URL
POSTHOG_TOKEN
GOOGLE_CLIENT_SECRET
R2_ACCESS_KEY
R2_ACCESS_SECRET_KEY
RESEND_API_KEY
# variables
APP_ENV
APP_EMAIL
BETTER_AUTH_URL
DOCKER_USERNAME
DROPLET_IP
GOOGLE_CLIENT_ID
R2_ACCOUNT_ID
R2_BUCKET_NAME
```

todo:
need to log actual response body
make sure golang logger works.
