# emptty
Dead simple Display Manager running in CLI as TTY login, that starts .xinitrc or .winitrc.

### Configuration
Configuration is handled via environment variables.

`TTY_NUMBER` TTY, where emptty will start.

`DEFAULT_USER` Preselected user, if AUTOLOGIN is enabled, this user is logged in.

`AUTOLOGIN` Enables Autologin, if DEFAULT_USER is defined. Possible values are "true" or "false".

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
- xorg (optional)
- xauth (required for xorg)
- mcookie (required for xorg)
- wayland (optional)

---
#### TODO:
- [x] PAM
- [x] Xorg support
- [x] runit service
- [x] Common failure handling
- [x] Configuration
- [x] Wayland
- [x] Autologin/Preset username
- [ ] systemd service
