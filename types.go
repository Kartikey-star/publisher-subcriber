package main

type HelmRequest struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	ChartUrl  string                 `json:"chart_url"`
	Values    map[string]interface{} `json:"values"`
	Version   int                    `json:"version"`
}

type HelmInstallations struct {
	Charts []HelmRequest `json:"charts"`
}

type HelmUninstallations struct {
	HelmReleases []HelmReleases `json:"releases"`
}

type HelmReleases struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
