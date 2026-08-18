# templ core

```bash
# sqlc dep
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
# migrate dep
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

![templ](./public/templ.gif)

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

In addition to env above
you will need docker and tailscale credential in github env

<img width="1336" height="262" alt="Screenshot_20260817_212922-1" src="https://github.com/user-attachments/assets/1d5d31bb-898e-4220-8e3a-9cb54dce0852" />

