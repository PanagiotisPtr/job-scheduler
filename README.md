# Job Scheduler
This is a microservices that manages kubernetes cron jobs. It is responsible for creating, starting, stopping and deleting them.
The service will look for cronjob files in GitHub repos (passed in the config) and store them locally. It then allows any of these
parsed cron jobs to be ran in the cluser.

Here's an overview of the endpoints available:

- List all available cron jobs
```
GET /static/jobs
```

- Show cron job config
```
GET /static/jobs/{cronJobName}
```

- Show cron jobs that are running in the cluster
```
GET /cluster/jobs
```

- Start a cron job
```
PATCH /cluster/jobs/{cronJobName}/start
```

- Stop a cron job
```
PATCH /cluster/jobs/{cronJobName}/stop
```

# Configuration
The configuration file is pulled by the service from a URL. That URL can be from an S3 bucket or any other service accessible to the job-scheduler
