# Troubleshooting

## Compilation problems

### gcc not found
### cannot find -lstdc++

This is usually an error on linux systems.  If you receive errors of this type your computer is missing some files to allow `ethdo` to build.  To resolve this run the following command:

```sh
sudo apt install build-essential libstdc++6
```

and then try to install `ethdo` again.

