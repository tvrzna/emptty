# emptty - Samples

## \~/.config/emptty or \~/.emptty as config
In your .config folder you have to create 'emptty' file or in your home folder you have to create `.emptty` file. If `environment` is not defined, it assumes xorg.

#### Xorg session
```
Name=Custom Optional Name
Exec=/usr/bin/openbox-session
Environment=xorg
```

#### Wayland session
```
Name=Custom Optional Name
Exec=/usr/bin/sway
Environment=wayland
```

## \~/.config/emptty or \~/.emptty as script
In your .config folder you have to create 'emptty' file or in your home folder you have to create `.emptty` file. This file needs to have execution permission (`chmod +x ~/.config/emptty` or `chmod +x ~/.emptty`).
```
#!/bin/sh
Environment=xorg

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

#### User-specific
Create folder custom-sessions as super user
```
$ mkdir -p ~/.config/emptty-custom-sessions/
```

#### System-wide
Create folder custom-sessions as super user
```
$ sudo mkdir -p /etc/emptty/custom-sessions
```

In these folders you can paste your desktop files. If `environment` is not defined, it assumes xorg.

### Xorg session
sowm.desktop

```
Name=sowm
Exec=/usr/bin/sowm
Environment=xorg
```


### Wayland session
sway.desktop

```
Name=My custom Sway
Exec=/usr/bin/sway
Environment=wayland
```