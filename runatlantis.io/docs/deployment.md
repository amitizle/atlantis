# Production-Ready Deployment
[[toc]]
## Install Terraform
`terraform` needs to be in the `$PATH` for Atlantis.
Download from https://www.terraform.io/downloads.html
```bash
unzip path/to/terraform_*.zip -d /usr/local/bin
```
Check that it's in your `$PATH`
```
$ terraform version
Terraform v0.10.0
```
If you want to use a different version of Terraform see [Terraform Versions](#terraform-versions)

## Hosting Atlantis
Atlantis needs to be hosted somewhere that github.com/gitlab.com or your GitHub/GitLab Enterprise installation can reach. Developers in your organization also need to be able to access Atlantis to view the UI and to delete locks.

By default Atlantis runs on port `4141`. This can be changed with the `--port` flag.

## Add GitHub Webhook
Once you've decided where to host Atlantis you can add it as a Webhook to GitHub.
If you already have a GitHub organization we recommend installing the webhook at the **organization level** rather than on each repository, however both methods will work.

::: tip
If you're not sure if you have a GitHub organization see https://help.github.com/articles/differences-between-user-and-organization-accounts/
:::

If you're installing on the organization, navigate to your organization's page and click **Settings**.
If installing on a single repository, navigate to the repository home page and click **Settings**.
- Select **Webhooks** or **Hooks** in the sidebar
- Click **Add webhook**
- set **Payload URL** to `http://$URL/events` where `$URL` is where Atlantis is hosted. **Be sure to add `/events`**
- set **Content type** to `application/json`
- set **Secret** to a random key (https://www.random.org/strings/). You'll need to pass this value to the `--gh-webhook-secret` flag when you start Atlantis
- select **Let me select individual events**
- check the boxes
	- **Pull request reviews**
	- **Pushes**
	- **Issue comments**
	- **Pull requests**
- leave **Active** checked
- click **Add webhook**

## Add GitLab Webhook
If you're using GitLab, navigate to your project's home page in GitLab
- Click **Settings > Integrations** in the sidebar
- set **URL** to `http://$URL/events` where `$URL` is where Atlantis is hosted. **Be sure to add `/events`**
- set **Secret Token** to a random key (https://www.random.org/strings/). You'll need to pass this value to the `--gitlab-webhook-secret` flag when you start Atlantis
- check the boxes
    - **Push events**
    - **Comments**
    - **Merge Request events**
- leave **Enable SSL verification** checked
- click **Add webhook**

## Create a GitHub Token
We recommend creating a new user in GitHub named **atlantis** that performs all API actions, however you can use any user.

**NOTE: The Atlantis user must have "Write permissions" (for repos in an organization) or be a "Collaborator" (for repos in a user account) to be able to set commit statuses:**
![Atlantis status](./images/status.png)

Once you've created the user (or have decided to use an existing user) you need to create a personal access token.
- follow [https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token)
- copy the access token

## Create a GitLab Token
We recommend creating a new user in GitLab named **atlantis** that performs all API actions, however you can use any user.
Once you've created the user (or have decided to use an existing user) you need to create a personal access token.
- follow [https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token](https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token)
- create a token with **api** scope
- copy the access token

## Start Atlantis
Now you're ready to start Atlantis!

If you're using GitHub, run:
```
atlantis server --atlantis-url $URL --gh-user $USERNAME --gh-token $TOKEN --gh-webhook-secret $SECRET
```

If you're using GitHub Enterprise, run:
```
HOSTNAME=YOUR_GITHUB_ENTERPRISE_HOSTNAME # ex. github.runatlantis.io, without the scheme
atlantis server --atlantis-url $URL --gh-user $USERNAME --gh-token $TOKEN --gh-webhook-secret $SECRET --gh-hostname $HOSTNAME
```

If you're using GitLab, run:
```
atlantis server --atlantis-url $URL --gitlab-user $USERNAME --gitlab-token $TOKEN --gitlab-webhook-secret $SECRET
```

If you're using GitLab Enterprise, run:
```
HOSTNAME=YOUR_GITLAB_ENTERPRISE_HOSTNAME # ex. gitlab.runatlantis.io, without the scheme
atlantis server --atlantis-url $URL --gitlab-user $USERNAME --gitlab-token $TOKEN --gitlab-webhook-secret $SECRET --gitlab-hostname $HOSTNAME
```

- `$URL` is the URL that Atlantis can be reached at
- `$USERNAME` is the GitHub/GitLab username you generated the token for
- `$TOKEN` is the access token you created. If you don't want this to be passed in as an argument for security reasons you can specify it in a config file (see [Configuration](#configuration)) or as an environment variable: `ATLANTIS_GH_TOKEN` or `ATLANTIS_GITLAB_TOKEN`
- `$SECRET` is the random key you used for the webhook secret. If you don't want this to be passed in as an argument for security reasons you can specify it in a config file (see [Configuration](#configuration)) or as an environment variable: `ATLANTIS_GH_WEBHOOK_SECRET` or `ATLANTIS_GITLAB_WEBHOOK_SECRET`

Atlantis is now running!
**We recommend running it under something like Systemd or Supervisord.**

## Docker
Atlantis also ships inside a docker image. Run the docker image:

```bash
docker run runatlantis/atlantis:latest server <required options>
```

### Usage
If you need to modify the Docker image that we provide, for instance to add a specific version of Terraform, you can do something like this:

* Create a custom docker file
```bash
vim Dockerfile-custom
```

```dockerfile
FROM runatlantis/atlantis

# copy a terraform binary of the version you need
COPY terraform /usr/local/bin/terraform
```

* Build docker image

```bash
docker build -t {YOUR_DOCKER_ORG}/atlantis-custom -f Dockerfile-custom .
```

* Run docker image

```bash
docker run {YOUR_DOCKER_ORG}/atlantis-custom server --gh-user=GITHUB_USERNAME --gh-token=GITHUB_TOKEN
```

## Kubernetes
Atlantis can be deployed into Kubernetes as a
[Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
or as a [Statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) with persistent storage.

StatefulSet is recommended because Atlantis stores its data on disk and so if your Pod dies
or you upgrade Atlantis, you won't lose the data. On the other hand, the only data that
Atlantis has right now is any plans that haven't been applied and Atlantis locks. If
Atlantis loses that data, you just need to run `atlantis plan` again so it's not the end of the world.

Regardless of whether you choose a Deployment or StatefulSet, first create a Secret with the webhook secret and access token:
```
echo -n "yourtoken" > token
echo -n "yoursecret" > webhook-secret
kubectl create secret generic atlantis-vcs --from-file=token --from-file=webhook-secret
```

Next, edit the manifests below as follows:
1. Replace `<VERSION>` in `image: runatlantis/atlantis:<VERSION>` with the most recent version from https://github.com/runatlantis/atlantis/releases/latest.
    * NOTE: You never want to run with `:latest` because if your Pod moves to a new node, Kubernetes will pull the latest image and you might end
up upgrading Atlantis by accident!
2. Replace `value: github.com/yourorg/*` under `name: ATLANTIS_REPO_WHITELIST` with the whitelist pattern
for your Terraform repos. See [--repo-whitelist](/docs/security.html#repo-whitelist) for more details.
3. If you're using GitHub:
    1. Replace `<YOUR_GITHUB_USER>` with the username of your Atlantis GitHub user without the `@`.
    2. Delete all the `ATLANTIS_GITLAB_*` environment variables.
4. If you're using GitLab:
    1. Replace `<YOUR_GITLAB_USER>` with the username of your Atlantis GitLab user without the `@`.
    2. Delete all the `ATLANTIS_GH_*` environment variables.

### StatefulSet Manifest
<details>
 <summary>Show...</summary>

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: atlantis
spec:
  serviceName: atlantis
  replicas: 1
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      partition: 0
  selector:
    matchLabels:
      app: atlantis
  template:
    metadata:
      labels:
        app: atlantis
    spec:
      securityContext:
        fsGroup: 1000 # Atlantis group (1000) read/write access to volumes.
      containers:
      - name: atlantis
        image: runatlantis/atlantis:v<VERSION> # 1. Replace <VERSION> with the most recent release.
        env:
        - name: ATLANTIS_REPO_WHITELIST
          value: github.com/yourorg/* # 2. Replace this with your own repo whitelist.

        ### GitHub Config ###
        - name: ATLANTIS_GH_USER
          value: <YOUR_GITHUB_USER> # 3i. If you're using GitHub replace <YOUR_GITHUB_USER> with the username of your Atlantis GitHub user without the `@`.
        - name: ATLANTIS_GH_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GH_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End GitHub Config ###

        ### GitLab Config ###
        - name: ATLANTIS_GITLAB_USER
          value: <YOUR_GITLAB_USER> # 4i. If you're using GitLab replace <YOUR_GITLAB_USER> with the username of your Atlantis GitLab user without the `@`.
        - name: ATLANTIS_GITLAB_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GITLAB_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End GitLab Config ###

        - name: ATLANTIS_DATA_DIR
          value: /atlantis
        - name: ATLANTIS_PORT
          value: "4141" # Kubernetes sets an ATLANTIS_PORT variable so we need to override.
        volumeMounts:
        - name: atlantis-data
          mountPath: /atlantis
        ports:
        - name: atlantis
          containerPort: 4141
        resources:
          requests:
            memory: 256Mi
            cpu: 100m
          limits:
            memory: 256Mi
            cpu: 100m
        livenessProbe:
          # We only need to check every 60s since Atlantis is not a
          # high-throughput service.
          periodSeconds: 60
          httpGet:
            path: /healthz
            port: 4141
            # If using https, change this to HTTPS
            scheme: HTTP
        readinessProbe:
          periodSeconds: 60
          httpGet:
            path: /healthz
            port: 4141
            # If using https, change this to HTTPS
            scheme: HTTP
  volumeClaimTemplates:
  - metadata:
      name: atlantis-data
    spec:
      accessModes: ["ReadWriteOnce"] # Volume should not be shared by multiple nodes.
      resources:
        requests:
          # The biggest thing Atlantis stores is the Git repo when it checks it out.
          # It deletes the repo after the pull request is merged.
          storage: 5Gi
---
apiVersion: v1
kind: Service
metadata:
  name: atlantis
spec:
  ports:
  - name: atlantis
    port: 80
    targetPort: 4141
  selector:
    app: atlantis
```
</details>


### Deployment Manifest
<details>
 <summary>Show...</summary>

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: atlantis
  labels:
    app: atlantis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: atlantis
  template:
    metadata:
      labels:
        app: atlantis
    spec:
      containers:
      - name: atlantis
        image: runatlantis/atlantis:v<VERSION> # 1. Replace <VERSION> with the most recent release.
        env:
        - name: ATLANTIS_REPO_WHITELIST
          value: github.com/yourorg/* # 2. Replace this with your own repo whitelist.

        ### GitHub Config ###
        - name: ATLANTIS_GH_USER
          value: <YOUR_GITHUB_USER> # 3i. If you're using GitHub replace <YOUR_GITHUB_USER> with the username of your Atlantis GitHub user without the `@`.
        - name: ATLANTIS_GH_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GH_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End GitHub Config ###

        ### GitLab Config ###
        - name: ATLANTIS_GITLAB_USER
          value: <YOUR_GITLAB_USER> # 4i. If you're using GitLab replace <YOUR_GITLAB_USER> with the username of your Atlantis GitLab user without the `@`.
        - name: ATLANTIS_GITLAB_TOKEN
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: token
        - name: ATLANTIS_GITLAB_WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: atlantis-vcs
              key: webhook-secret
        ### End GitLab Config ###

        - name: ATLANTIS_PORT
          value: "4141" # Kubernetes sets an ATLANTIS_PORT variable so we need to override.
        ports:
        - name: atlantis
          containerPort: 4141
        resources:
          requests:
            memory: 256Mi
            cpu: 100m
          limits:
            memory: 256Mi
            cpu: 100m
        livenessProbe:
          # We only need to check every 60s since Atlantis is not a
          # high-throughput service.
          periodSeconds: 60
          httpGet:
            path: /healthz
            port: 4141
            # If using https, change this to HTTPS
            scheme: HTTP
        readinessProbe:
          periodSeconds: 60
          httpGet:
            path: /healthz
            port: 4141
            # If using https, change this to HTTPS
            scheme: HTTP
---
apiVersion: v1
kind: Service
metadata:
  name: atlantis
spec:
  ports:
  - name: atlantis
    port: 80
    targetPort: 4141
  selector:
    app: atlantis
```
</details>

### Routing and SSL
The manifests above create a Kubernetes `Service` of type `ClusterIP` which isn't accessible outside your cluster.
Depending on how you're doing routing into Kubernetes, you may want to use a `LoadBalancer` so that Atlantis is accessible
to GitHub/GitLab and your internal users.

If you want to add SSL you can use something like https://github.com/jetstack/cert-manager to generate SSL
certs and mount them into the Pod. Then set the `ATLANTIS_SSL_CERT_FILE` and `ATLANTIS_SSL_KEY_FILE` environment variables to enable SSL.
You could also set up SSL at your LoadBalancer.

## AWS Fargate

If you'd like to run Atlantis on [AWS Fargate](https://aws.amazon.com/fargate/) check out the Atlantis module on the Terraform Module Registry: https://registry.terraform.io/modules/terraform-aws-modules/atlantis/aws

## Testing Out Atlantis on GitHub

If you'd like to test out Atlantis before running it on your own repositories you can fork our example repo.

- Fork https://github.com/runatlantis/atlantis-example
- If you didn't add the Webhook as to your organization add Atlantis as a Webhook to the forked repo (see [Add GitHub Webhook](#add-github-webhook))
- Now that Atlantis can receive events you should be able to comment on a pull request to trigger Atlantis. Create a pull request
	- Click **Branches** on your forked repo's homepage
	- click the **New pull request** button next to the `example` branch
	- Change the `base` to `{your-repo}/master`
	- click **Create pull request**
- Now you can test out Atlantis
	- Create a comment `atlantis help` to see what commands you can run from the pull request
	- `atlantis plan` will run `terraform plan` behind the scenes. You should see the output commented back on the pull request. You should also see some logs show up where you're running `atlantis server`
	- `atlantis apply` will run `terraform apply`. Since our pull request creates a `null_resource` (which does nothing) this is safe to do.


