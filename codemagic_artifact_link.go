package codemagic_notifier

type CodeMagicArtifactLink struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Url         string `json:"url"`
	Md5         string `json:"md5"`
	VersionName string `json:"versionName"`
	BundleId    string `json:"bundleId"`
}
