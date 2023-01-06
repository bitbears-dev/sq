# 💻 sq: jq for system info 🤓

## Examples

```
$ sq /proc/vmstat
{
  "nr_free_pages": 287330,
  "nr_zone_inactive_anon": 4157354
  "nr_zone_active_anon": 725726
  "nr_zone_inactive_file": 975890
  "nr_zone_active_file": 960405
  "nr_zone_unevictable": 670303
  "nr_zone_write_pending": 712
  "nr_mlock": 3223
  "nr_bounce": 0
  "nr_zspages": 0
  "nr_free_cma": 0
  "numa_hit": 5267104191
  "numa_miss": 0
  "numa_foreign": 0
  "numa_interleave": 3053
  "numa_local": 5267031177
  "numa_other": 0
  ...
}
```


```
$ sq .nr_free_pages /proc/vmstat
287330
```


## Install

By running one of the following commands, the latest version of `sq` command will be installed.

If you have a permission to write a file into `/usr/local/bin` directory (e.g. you are `root` user), please run the command below:

```shell
curl -fsSL https://raw.githubusercontent.com/bitbears-dev/sq/master/install.sh | bash
```

If you do not have a permission to write a file into `/usr/local/bin` directory, please run either of the following commands.

If you are in sudoers and want to install `sq` command to `/usr/local/bin`:

```shell
curl -fsSL https://raw.githubusercontent.com/bitbears-dev/sq/master/install.sh | sudo bash
```

or

If you are not in sudoers or want to install `sq` command to other directory e.g. `$HOME/bin`:

```shell
mkdir -p "$HOME/bin"
curl -fsSL https://raw.githubusercontent.com/bitbears-dev/sq/master/install.sh | BINDIR="$HOME/bin" bash
```

You can change `"$HOME/bin"` in the command above to wherever you want.

If you want to upgrade the `sq` command, you can just run the same command you used to install `sq` again.

If you want to uninstall the `sq` command, you can just remove `sq` executable file you have installed.

If the commands above did not work well, or if you want to install older version of `sq` command, you can download a package file that match the environment of the target from [Releases page](https://github.com/bitbears-dev/sq/releases), unpack it, and place the executable file in the directory where included in `PATH`.



## Reference

### Supported files

<details>
<summary>Linux</summary>

  <details>
  <summary>/proc</summary>

  </details>

</details>


# Development

## How to release

```
make build-for-release ver=x.y.z
make package ver=x.y.z
make release ver=x.y.z
```
