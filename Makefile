DISTFILE=emptty

clean:
	@echo "Cleaning..."
	@rm -f ${DISTFILE}
	@echo "Done"

build:
	@echo "Building..."
	@go get github.com/bgentry/speakeasy
	@go get github.com/msteinert/pam
	@go build -o ${DISTFILE}
	@echo "Done"

install:
	@echo "Installing..."
	@install -DZs ${DISTFILE} -m 755 -t ${DESTDIR}/usr/bin
	@install -DZ res/pam -m 644 -T ${DESTDIR}/etc/pam.d/${DISTFILE}
	@echo "Done"

install-config:
	@echo "Installing config..."
	@install -DZ res/conf -m 755 -T ${DESTDIR}/etc/${DISTFILE}/conf
	@echo "Done"

install-runit:
	@echo "Installing runit service..."
	@install -DZ res/runit-run -m 755 -T ${DESTDIR}/etc/sv/${DISTFILE}/run
	@install -DZ res/runit-finish -m 755 -T ${DESTDIR}/etc/sv/${DISTFILE}/finish
	@echo "Done"


install-systemd:
	@echo "Installing systemd service..."
	@install -DZ res/systemd-service -m 755 -T ${DESTDIR}/usr/lib/systemd/system/${DISTFILE}.service
	@echo "Done"

uninstall:
	@echo "Uninstalling..."
	@rm -rf ${DESTDIR}/etc/sv/${DISTFILE}
	@rm -f ${DESTDIR}/usr/lib/systemd/system/${DISTFILE}.service
	@rm -f ${DESTDIR}/etc/pam.d/emptty
	@rm -rf ${DESTDIR}/usr/bin/${DISTFILE}
	@echo "Done"
