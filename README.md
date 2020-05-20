# emptty
Dead simple Display Manager running in CLI as TTY login, that starts Xorg or Wayland.

### Configuration
Configuration is handled via environment variables.

`TTY_NUMBER` TTY, where emptty will start.

`DEFAULT_USER` Preselected user, if AUTOLOGIN is enabled, this user is logged in.

`AUTOLOGIN` Enables Autologin, if DEFAULT_USER is defined. Possible values are "true" or "false".
__NOTE:__ to enable autologin DEFAULT_USER must be in group nopasswdlogin, otherwise user will NOT be authorized.

`ENVIRONMENT` Selects, which environment will be used. Possible values are "xorg" or "wayland".
- If "xorg" is selected, it expects to have prepared .xinitrc file with +x at home folder.
- If "wayland" is selected, it expects to have prepared .winitrc file with +x at home folder.

### .xinitrc sample
```
#!/bin/sh
export LANG=en_US.UTF-8
exec dbus-launch i3
```

### .winitrc sample
```
#!/bin/sh
export LANG=en_US.UTF-8
exec dbus-launch sway
```

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
