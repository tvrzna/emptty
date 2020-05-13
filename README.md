# emptty
Dead simple Display Manager running in CLI as TTY login, that starts .xinitrc or .winitrc.

## Build dependencies
- go
- gcc
- pam-devel

## Dependencies
- pam
- xorg (optional)
- xauth (required for xorg)
- mcookie (required for xorg)
- wayland (optional)

## .xinitrc sample
```
#!/bin/sh
export LANG=en_US.UTF-8
exec dbus-launch i3
```

## .winitrc sample
```
#!/bin/sh
export LANG=en_US.UTF-8
exec dbus-launch sway
```

---
### TODO:
- [x] PAM
- [x] Xorg support
- [x] runit service
- [x] Common failure handling
- [x] Configuration/Arguments
- [x] Wayland
- [x] Autologin/Preset username
- [ ] systemd service
