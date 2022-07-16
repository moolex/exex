# exex

Similar to use sh to run program with extend args but more powerful

1. forward signals to child process
2. try close child process when stdin pipe closed

## Usage

Before:
```sh
#!/bin/sh
podman -c local $@
```

After:
```sh
ln -s exex docker
echo "podman" > docker.alias
echo "-c local" > docker.args
```

and you can now use `docker` as `podman -c local`
