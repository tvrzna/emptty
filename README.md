# emptty
Dead simple Display Manager running in CLI as TTY login, that starts Xorg or Wayland.

### Configuration

##### /etc/emptty/conf
Default startup configuration. On each change it requires to restart emptty.

`TTY_NUMBER` TTY, where emptty will start.

`DEFAULT_USER` Preselected user, if AUTOLOGIN is enabled, this user is logged in.

`AUTOLOGIN` Enables Autologin, if DEFAULT_USER is defined. Possible values are "true" or "false".
__NOTE:__ to enable autologin DEFAULT_USER must be in group nopasswdlogin, otherwise user will NOT be authorized.

`LANG` defines locale for all users. Default value is "en_US.UTF-8"

##### ${HOME}/.emptty
Optional configuration file, that could be also handled as shell script. If is not presented, emptty shows selection of installed desktops.

`ENVIRONMENT` Selects, which environment should be defined for following command. Possible values are "xorg" and "wayland", "xorg" is default.

`COMMAND` Defines command to start Desktop Environment/Window Manager. This value does not need to be defined, if .emptty file is presented as shell script (with shebang at the start and execution permissions).

`LANG` Defines locale for logged user, has higner priority than LANG from global configuration

### Build dependencies
- go
- gcc
- pam-devel

### Dependencies
- pam
- xorg / xorg-server (optional)
- xauth / xorg-xauth (required for xorg)
- mcookie (required for xorg)
- wayland (optional)

## Build & install
- `make clean` to cleanup already built binary.
- `make install` to install binary and pam module.
- `make install-config` to create default conf file in /etc/emptty/.
- `make install-runit` to install as runit service
- `make install-openrc` to install as openrc service
- `make install-systemd` to install as systemd service.
