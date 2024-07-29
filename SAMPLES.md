# emptty - Samples

## \~/.config/emptty or \~/.emptty as init script
In your `.config` folder you have to create 'emptty' file or in your home folder you have to create `.emptty` file.

This variant allows to treat your script in similar way as your `.xinitrc`, however this is common to both Xorg and Wayland. The magic option is `Selection=true` or `Selection=auto`. You can define your own environmental variables and keep the possibility to select any desktop.

As it is mentioned in [README](README.md), no `.profile` scripts are sourced by default. However following scripts contains few examples, how it could be done inside `emptty` file.

#### Script with sourced `/etc/profile`
```
#!/bin/sh
Selection=true

xrandr --output eDP1 --mode 1920x1080
xrdb -merge ~/.Xresources

# source /etc/profile does not have any effect
. /etc/profile
. ~/.bashrc

export BROWSER=firefox
export EDITOR=vim

exec dbus-launch $@
```

#### Script with sourced `/etc/profile` using LoginShell
```
#!/bin/sh
Selection=true
LoginShell=/bin/bash --login

# /etc/profile is sourced by using non-interactive shell

exec dbus-launch $@
```

#### Script with fish support in LoginShell
Emptty supports simplified fish support, if shebang is set to fish, properties could be set in its basic way.
```
#!/bin/fish
set Selection true
set LoginShell /bin/fish --login

exec dbus-launch $argv
```

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

# source /etc/profile does not have any effect
. /etc/profile
. ~/.bashrc

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
$ mkdir -p ~/.config/emptty/custom-sessions
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