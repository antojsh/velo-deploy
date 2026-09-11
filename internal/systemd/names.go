package systemd

const watcherUnit = "velo-deploy-watcher.service"

func UnitName(appName string) string {
	return "velo-" + appName + ".service"
}

func LegacyUnitName(appName string) string {
	return "deploy-" + appName + ".service"
}

func AppUser(appName string) string {
	return "velo-" + appName
}

func ResolveUnit(appName string) string {
	return resolveExistingUnit(appName)
}

func unitCandidates(appName string) []string {
	return []string{UnitName(appName), LegacyUnitName(appName)}
}
