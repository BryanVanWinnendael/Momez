# 📷 Momez - Capture and save your favorite moments

## Features
- Create an account
- Share your photos with a caption to the world
- Discover moments with a search of a caption
- Favorite moments
- Choose your profile picture
- Edit or delete posts
- Visit others' profiles

## Setup

## Firebase 
- Create a [firebase db](https://firebase.google.com/)
- Setup A Realtime Database
- Setup Storage

In firebase Realtime Database > Rules, add the rule:
```json
"posts": {
    ".indexOn": ["id", "username"]
  }
```

### Firebase serviceAccountKey
Download the serviceAccountKey.json from
[firebase admin](https://firebase.google.com/docs/admin/setup#initialize_the_sdk_in_non-google_environments) and put it in the root dir.

## .env
Create a .env in the root dir with

```json
DATABASE_URL="url"
BUCKET_NAME="bucketname"
SECRET_KEY="secretkey"
```
DATABASE_URL and BUCKET_NAME can be found when going to Project settings > Your apps in firebase.
If you don't see it, create an app.
For the SECRET_KEY any value will do. This value will be used by the cookieStore.

## Run with docker

### Requirements
- [Docker](https://docs.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Run it
```
docker compose up -d
```

## Run without docker

### Requirements
- [Go](https://go.dev/doc/install)
- [Air](https://github.com/cosmtrek/air)


### Windows/Linux/MacOS
In the .air.toml file change the "./tmp/main" to "./tmp/main.exe" for windows, for Linux and MacOS no changes have to be made.

### Run it
``` bash
air
```

## Making changes

### Requirements
- [Tailwind CSS CLI](https://tailwindcss.com/docs/installation)

### Tailwind CSS
When making changes to the code with tailwind make sure to run.
```bash
tailwindcss -i css/input.css -o css/output.css --watch
```
