DISTFILE=emptty
BUILD_VERSION=`git describe --tags`
GOVERSION=`go version | grep -Eo 'go[0-9]+\.[0-9]+'`

ifdef TAGS
	TAGS_ARGS = -tags ${TAGS}
endif

test:
	@echo "Testing..."
	@go test -coverprofile cover.out ${TAGS_ARGS} ./...
	@echo "Done"

clean:
	@echo "Cleaning..."
	@rm -f dist/${DISTFILE}
	@rm -f dist/emptty.1.gz
	@rm -rf dist
	@echo "Done"

build:
	@echo "Building${TAGS_ARGS}..."
	@mkdir -p dist
	@go build ${TAGS_ARGS} -o dist/${DISTFILE} -ldflags "-X github.com/tvrzna/emptty/src.buildVersion=${BUILD_VERSION}" -buildvcs=false
	@gzip -cn res/emptty.1 > dist/emptty.1.gz
	@echo "Done"

install:
	@echo "Installing..."
	@install -DZs dist/${DISTFILE} -m 755 -t ${DESTDIR}/usr/bin
	@echo "Done"

install-config:
	@echo "Installing config..."
	@install -DZ res/conf -m 644 -T ${DESTDIR}/etc/${DISTFILE}/conf
	@echo "Done"

install-manual:
	@echo "Installing manual..."
	@install -D dist/emptty.1.gz -t ${DESTDIR}/usr/share/man/man1
	@echo "Done"

install-motd-gen:
	@echo "Installing motd-gen.sh..."
	@install -DZ res/motd-gen.sh -m 744 -t ${DESTDIR}/etc/${DISTFILE}/
	@echo "Done"

install-pam:
	@echo "Installing pam file..."
	@install -DZ res/pam -m 644 -T ${DESTDIR}/etc/pam.d/${DISTFILE}
	@echo "Done"

install-pam-fedora:
	@echo "Installing pam-fedora file..."
	@install -DZ res/pam-fedora -m 644 -T ${DESTDIR}/etc/pam.d/${DISTFILE}
	@echo "Done"

install-pam-suse:
	@echo "Installing pam-suse file..."
	@install -DZ res/pam-suse -m 644 -T ${DESTDIR}/etc/pam.d/${DISTFILE}
	@echo "Done"

install-runit:
	@echo "Installing runit service..."
	@install -DZ res/runit-run -m 755 -T ${DESTDIR}/etc/sv/${DISTFILE}/run
	@echo "Done"

install-runit-artix:
	@echo "Installing Artix runit service..."
	@install -DZ res/runit-run -m 755 -T ${DESTDIR}/etc/runit/sv/${DISTFILE}/run
	@echo "Done"

install-systemd:
	@echo "Installing systemd service..."
	@install -DZ res/systemd-service -m 644 -T ${DESTDIR}/usr/lib/systemd/system/${DISTFILE}.service
	@echo "Done"

install-openrc:
	@echo "Installing OpenRC service..."
	@install -DZ res/openrc-service -m 755 -T ${DESTDIR}/etc/init.d/${DISTFILE}
	@echo "Done"

install-s6:
	@echo "Installing S6 service..."
	@install -DZ res/s6-dependencies -m 644 -T ${DESTDIR}/etc/s6/sv/${DISTFILE}/dependencies
	@install -DZ res/s6-type -m 644 -T ${DESTDIR}/etc/s6/sv/${DISTFILE}/type
	@install -DZ res/s6-run -m 755 -T ${DESTDIR}/etc/s6/sv/${DISTFILE}/run
	@echo "Done. Please recompile your S6 database."

install-dinit:
	@echo "Installing dinit service..."
	@install -DZ res/dinit-service -m 644 -T ${DESTDIR}/etc/dinit.d/${DISTFILE}
	@install -DZ res/dinit-script -m 755 -T ${DESTDIR}/etc/dinit.d/scripts/${DISTFILE}
	@echo "Done"

install-all: install install-manual install-pam

uninstall:
	@echo "Uninstalling..."
	@rm -rf ${DESTDIR}/etc/sv/${DISTFILE}
	@rm -rf ${DESTDIR}/etc/runit/sv/${DISTFILE}
	@rm -f ${DESTDIR}/usr/lib/systemd/system/${DISTFILE}.service
	@rm -f ${DESTDIR}/etc/init.d/${DISTFILE}
	@rm -f ${DESTDIR}/usr/share/man/man1/emptty.1.gz
	@rm -f ${DESTDIR}/etc/pam.d/emptty
	@rm -rf ${DESTDIR}/etc/s6/sv/${DISTFILE}
	@rm -rf ${DESTDIR}/usr/bin/${DISTFILE}
	@rm -rf ${DESTDIR}/etc/dinit.d/${DISTFILE}
	@rm -rf ${DESTDIR}/etc/dinit.d/scripts/${DISTFILE}
	@echo "Done"
