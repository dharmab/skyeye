package types

// ServerSetting identifies a server-side setting.
type ServerSetting string

const (
	ServerPort                    ServerSetting = "SERVER_PORT"
	CoalitionAudioSecurity        ServerSetting = "COALITION_AUDIO_SECURITY"
	SpectatorsAudioDisabled       ServerSetting = "SPECTATORS_AUDIO_DISABLED"
	ClientExportEnabled           ServerSetting = "CLIENT_EXPORT_ENABLED"
	LOSEnabled                    ServerSetting = "LOS_ENABLED"
	DistanceEnabled               ServerSetting = "DISTANCE_ENABLED"
	IRLRadioTX                    ServerSetting = "IRL_RADIO_TX"
	IRLRadioRXInterference        ServerSetting = "IRL_RADIO_RX_INTERFERENCE"
	IRLRadioStatic                ServerSetting = "IRL_RADIO_STATIC"
	RadioExpansion                ServerSetting = "RADIO_EXPANSION"
	ExternalAWACSMode             ServerSetting = "EXTERNAL_AWACS_MODE"
	ExternalAWACSModeBluePassword ServerSetting = "EXTERNAL_AWACS_MODE_BLUE_PASSWORD" // #nosec G101
	ExternalAWACSModeRedPassword  ServerSetting = "EXTERNAL_AWACS_MODE_RED_PASSWORD"  // #nosec G101
	ClientExportFilePath          ServerSetting = "CLIENT_EXPORT_FILE_PATH"
	CheckForBetaUpdates           ServerSetting = "CHECK_FOR_BETA_UPDATES"
	AllowRadioEncryption          ServerSetting = "ALLOW_RADIO_ENCRYPTION"
	TestFrequencies               ServerSetting = "TEST_FREQUENCIES"
	ShowTunedCount                ServerSetting = "SHOW_TUNED_COUNT"
	GlobalLobbyFrequencies        ServerSetting = "GLOBAL_LOBBY_FREQUENCIES"
	ShowTransmitterName           ServerSetting = "SHOW_TRANSMITTER_NAME"
	LotATCExportEnabled           ServerSetting = "LOTATC_EXPORT_ENABLED"
	LotATCExportPort              ServerSetting = "LOTATC_EXPORT_PORT"
	LotATCExportIP                ServerSetting = "LOTATC_EXPORT_IP"
	UPnPEnabled                   ServerSetting = "UPNP_ENABLED"
	RetransmissionNodeLimit       ServerSetting = "RETRANSMISSION_NODE_LIMIT"
	StrictRadioEncryption         ServerSetting = "STRICT_RADIO_ENCRYPTION"
	TransmissionLogEnabled        ServerSetting = "TRANSMISSION_LOG_ENABLED"
	TransmissionLogRetention      ServerSetting = "TRANSMISSION_LOG_RETENTION"
	RadioEffectOverride           ServerSetting = "RADIO_EFFECT_OVERRIDE"
	ServerIP                      ServerSetting = "SERVER_IP"
)
