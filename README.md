# exex

Similar to use sh to run program with extend args

before:
```sh
#!/bin/sh
podman -c local $@
```

after:
```sh
ln -s exex docker
echo "podman" > docker.alias
echo "-c local" > docker.args
```

and you can now use `docker` as `podman -c local`
