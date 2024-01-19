FROM archlinux:latest
RUN pacman -Syu --noconfirm && pacman -S make go sudo bash vim gcc pam util-linux xorg xorg-server --noconfirm
WORKDIR app/
COPY . ./
RUN useradd -m -U sysuser
RUN make && make install-all
ENTRYPOINT "/bin/sh"