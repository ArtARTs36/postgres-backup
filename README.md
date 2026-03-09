# postgres-backup

Backup PostgreSQL databases to S3.

## Usage with Swarm and CronJob

[CronJob for Swarm](https://github.com/crazy-max/swarm-cronjob)

```yaml
services:
  postgres-backup:
    image: artarts36/postgres-backup:latest
    secrets:
      - postgres18-backuper-s3-key-id
      - postgres18-backuper-s3-key-secret
      - postgres18-password
    environment:
      MAX_BACKUPS: 30

      STORAGE_S3_ENDPOINT: s3.cloud.ru
      STORAGE_S3_REGION: ru-central-1
      STORAGE_S3_ACCESS_KEY_FILE: /run/secrets/postgres-backuper-s3-key-id
      STORAGE_S3_SECRET_KEY_FILE: /run/secrets/postgres-backuper-s3-key-secret
      STORAGE_S3_BUCKET: postgres-backups
      STORAGE_S3_USE_SSL: 'true'

      PG_DATABASE: users, posts
      PG_USER: iuser
      PG_PASSWORD_FILE: /run/secrets/infra-postgres18-password
      PG_HOST: postgres18
      PG_PORT: 5432
    networks:
      - infra
    deploy:
      labels:
        - "swarm.cronjob.enable=true"
        - "swarm.cronjob.schedule=0 0 * * *"
        - "swarm.cronjob.skip-running=false"
      restart_policy:
        condition: none
```
