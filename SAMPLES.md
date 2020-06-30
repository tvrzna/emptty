# emptty - Samples

## \~/.emptty as config
In your home folder you have to create `.emptty` file. If `environment` is not defined, it assumes xorg.

#### Xorg session
```
command=/usr/bin/openbox-session
environment=xorg
```

#### Wayland session
```
command=/usr/bin/sway
environment=wayland
```

## \~/.emptty as script
In your home folder you have to create `.emptty` file. This file needs to have execution permission (`chmod +x ~/.emptty`).
```
#!/bin/sh
environment=xorg

exec dbus-launch i3
```

## \~/.xinitrc
In your home folder you have to create `.xinitrc` file. This file needs to have execution permission (`chmod +x ~/.xinitrc`).

```
#!/bin/sh

. ~/.xprofile
xrdb -merge ~/.Xresources
xmodmap ~/.Xmodmap

exec dbus-launch $@
```


## Custom sessions
Create folder custom-sessions as super user
```
$ sudo mkdir -p /etc/emptty/custom-sessions
```

In that folder you can paste your desktop files. If `environment` is not defined, it assumes xorg.

#### Xorg session
sowm.desktop

```
Name=sowm
Exec=/usr/bin/sowm
Environment=xorg
```


#### Wayland session
sway.desktop

```
Name=My custom Sway
Exec=/usr/bin/sway
Environment=wayland
```