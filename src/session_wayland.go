package src

// waylandSession defines structure for wayland
type waylandSession struct {
	*commonSession
}

// Starts no wayland carrier
func (w *waylandSession) startCarrier() {
}

// Gets -1 as carrier Pid
func (w *waylandSession) getCarrierPid() int {
	return -1
}

// Finishes no wayland carrier
func (w *waylandSession) finishCarrier() error {
	return nil
}
