package validations

import (
	api "k8s.io/api/core/v1"
)

#STIGPodSpec: api.#PodSpec & {
	containers?: [...#STIGContainer]
	initContainers?: [...#STIGContainer]
	securityContext!: #STIGPodSecurityContext

	hostPID?:     null | false
	hostNetwork?: null | false
	hostIPC?:     null | false
}

#STIGContainer: api.#Container & {
	ports?: [..._#ValidPort]
	securityContext!: #STIGSecurityContext

	envFrom?: secretRef?: null
	env?: [...api.#EnvVar & {
		valueFrom?: secretKeyRef?: null
	}]
}

#STIGPodSecurityContext: api.#PodSecurityContext & {
	runAsNonRoot!: true
	runAsUser?:    >0

	seccompProfile!: type!: "Localhost" | "RuntimeDefault"

	seLinuxOptions?: {
		user?: null
		role?: null
	}

	sysctls?: [..._#SafeSysctl]
}

#STIGSecurityContext: api.#SecurityContext & {
	allowPrivilegeEscalation!: false
	privileged?:               false
	readOnlyRootFilesystem!:   true
	runAsNonRoot?:             true
	seccompProfile?: {
		type!: "Localhost" | "RuntimeDefault"
	}
	capabilities!: api.#Capabilities & {
		drop?: ["ALL"]
		add?: [...(#AllowedCapabilities | error("capability not allowed found in 'add' list"))]
	}

	seLinuxOptions?: {
		user?: null
		role?: null
	}
}

#AllowedCapabilities: null | "AUDIT_WRITE" | "CHOWN" | "DAC_OVERRIDE" | "FOWNER" | "FSETID" | "KILL" | "MKNOD" |
	"NET_BIND_SERVICE" | "SETFCAP" | "SETGID" | "SETPCAP" | "SETUID" | "SYS_CHROOT"

#SysctlsAllowedNames: "kernel.shm_rmid_forced" | "net.ipv4.ip_local_port_range" | "net.ipv4.ip_unprivileged_port_start" | "net.ipv4.tcp_syncookies" |
	"net.ipv4.ping_group_range" | "net.ipv4.ip_local_reserved_ports" | "net.ipv4.tcp_keepalive_time" | "net.ipv4.tcp_fin_timeout" |
	"net.ipv4.tcp_keepalive_intvl" | "net.ipv4.tcp_keepalive_probes"

_#ValidEnvVarSource: api.#EnvVarSource & {
	secretKeyRef?: null | error("env vars must not get values from secrets")
}
_#ValidEnv: api.#EnvVar & {
	valueFrom?: null | _#ValidEnvVarSource
}

_#SafeSysctl: api.#Sysctl & {
	name?: #SysctlsAllowedNames
}

_#ValidPort: api.#ContainerPort & {
	hostPort?: 0
}
