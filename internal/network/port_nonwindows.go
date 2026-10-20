//go:build !windows

package network

func findPortOwnerOS(targetPort int) (int, string, string, error) {
	return 0, "", "", nil
}
