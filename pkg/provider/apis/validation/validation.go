package validation

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"

	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
)

// uuidRegex is a regex pattern for validating UUID format
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// keypairNameRegex is a regex pattern for validating keypair name format
// Pattern from STACKIT API: ^[A-Za-z0-9@._-]*$
var keypairNameRegex = regexp.MustCompile(`^[A-Za-z0-9@._-]*$`)

// emailRegex is a simple regex pattern for validating email format
// Basic validation: local-part@domain with reasonable character restrictions
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// machineTypeRegex is a regex pattern for validating machine type format
// Pattern: lowercase letter(s) followed by digits, dot, then more digits (e.g., c2i.2, m2i.8, g1a.8)
var machineTypeRegex = regexp.MustCompile(`^[a-z]+\d+[a-z]*\.\d+[a-z]*(\.[a-z]+\d+)*$`)

// regionRegex is a regex pattern for validating STACKIT region format
// Pattern: lowercase letters/digits (e.g., eu01, eu01)
var regionRegex = regexp.MustCompile(`^[a-z0-9]+$`)

// availabilityZoneRegex is a regex pattern for validating STACKIT availability zone format
// Pattern: lowercase letters/digits followed by digits, dash, then digit(s) (e.g., eu01-1, eu01-2)
var availabilityZoneRegex = regexp.MustCompile(`^[a-z0-9]+-\d+$`)

// labelKeyRegex validates Kubernetes label keys (must start/end with alphanumeric, can contain -, _, .)
// Maximum length: 63 characters
var labelKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9]([-a-zA-Z0-9_.]*[a-zA-Z0-9])?$`)

// labelValueRegex validates Kubernetes label values (must start/end with alphanumeric, can contain -, _, ., can be empty)
// Maximum length: 63 characters
var labelValueRegex = regexp.MustCompile(`^([a-zA-Z0-9]([-a-zA-Z0-9_.]*[a-zA-Z0-9])?)?$`)

// ValidateProviderSpecNSecret validates provider spec and secret to check if all fields are present and valid
//
//nolint:gocyclo,funlen//TODO:refactor
func ValidateProviderSpecNSecret(spec *api.ProviderSpec, secrets *corev1.Secret) []error {
	var errors []error

	// Validate Secret
	if secrets == nil {
		errors = append(errors, fmt.Errorf("secret is required"))
		return errors // Return early if secret is nil
	}

	projectID, ok := secrets.Data["project-id"]
	if !ok {
		errors = append(errors, fmt.Errorf("secret field 'project-id' is required"))
	} else if len(projectID) == 0 {
		errors = append(errors, fmt.Errorf("secret field 'project-id' cannot be empty"))
	} else if !isValidUUID(string(projectID)) {
		errors = append(errors, fmt.Errorf("secret field 'project-id' must be a valid UUID"))
	}

	// Validate serviceAccountKey (required for authentication)
	// ServiceAccount Key Flow: JSON string containing service account credentials and private key
	serviceAccountKey, ok := secrets.Data["serviceaccount.json"]
	if !ok {
		errors = append(errors, fmt.Errorf("secret field 'serviceaccount.json' is required"))
	} else if len(serviceAccountKey) == 0 {
		errors = append(errors, fmt.Errorf("secret field 'serviceaccount.json' cannot be empty"))
	} else if !isValidJSON(string(serviceAccountKey)) {
		errors = append(errors, fmt.Errorf("secret field 'serviceAccountKey' must be valid JSON (service account credentials)"))
	}

	// Validate region (required for SDK)
	if spec.Region == "" {
		errors = append(errors, fmt.Errorf("providerSpec.Region cannot be empty"))
	} else if !isValidRegion(spec.Region) {
		errors = append(errors, fmt.Errorf("providerSpec.Region has invalid format (expected format: eu01, eu02, etc.)"))
	}

	if spec.AvailabilityZone != "" && !isValidAvailabilityZone(spec.AvailabilityZone) {
		errors = append(errors, fmt.Errorf("providerSpec.availabilityZone has invalid format (expected format: eu01-1, eu01-2, etc.)"))
	}

	// Validate ProviderSpec
	if spec.MachineType == "" {
		errors = append(errors, fmt.Errorf("providerSpec.machineType is required"))
	} else if !isValidMachineType(spec.MachineType) {
		errors = append(errors, fmt.Errorf("providerSpec.machineType has invalid format (expected format: c2i.2, m2i.8, etc.)"))
	}

	// ImageID is required unless BootVolume.Source is specified
	hasBootVolumeSource := spec.BootVolume != nil && spec.BootVolume.Source != nil
	if spec.ImageID == "" && !hasBootVolumeSource {
		errors = append(errors, fmt.Errorf("providerSpec.imageId or bootVolume.source is required"))
	}
	// Validate ImageID format if specified
	if spec.ImageID != "" && !isValidUUID(spec.ImageID) {
		errors = append(errors, fmt.Errorf("providerSpec.imageId must be a valid UUID"))
	}

	// Validate Labels
	if spec.Labels != nil {
		for key, value := range spec.Labels {
			if len(key) > 63 {
				errors = append(errors, fmt.Errorf("providerSpec.labels key '%s' exceeds maximum length of 63 characters", key))
			}
			if !labelKeyRegex.MatchString(key) {
				errors = append(errors, fmt.Errorf("providerSpec.labels key '%s' has invalid format (must start/end with alphanumeric, can contain -, _, .)", key))
			}
			if len(value) > 63 {
				errors = append(errors, fmt.Errorf("providerSpec.labels value for key '%s' exceeds maximum length of 63 characters", key))
			}
			if !labelValueRegex.MatchString(value) {
				errors = append(errors, fmt.Errorf("providerSpec.labels value for key '%s' has invalid format (must start/end with alphanumeric, can contain -, _, ., can be empty)", key))
			}
		}
	}

	// Validate Networking
	if spec.Networking != nil {
		networkingErrors := validateNetworking(spec.Networking)
		errors = append(errors, networkingErrors...)
	}

	// Validate SecurityGroups
	if len(spec.SecurityGroups) > 0 {
		for i, sg := range spec.SecurityGroups {
			if sg == "" {
				errors = append(errors, fmt.Errorf("providerSpec.securityGroups[%d] cannot be empty", i))
			}
		}
	}

	// Validate BootVolume
	if spec.BootVolume != nil {
		bootVolumeErrors := validateBootVolume(spec.BootVolume)
		errors = append(errors, bootVolumeErrors...)
	}

	// Validate Volumes
	if len(spec.Volumes) > 0 {
		for i, volumeID := range spec.Volumes {
			if volumeID == "" {
				errors = append(errors, fmt.Errorf("providerSpec.volumes[%d] cannot be empty", i))
			} else if !isValidUUID(volumeID) {
				errors = append(errors, fmt.Errorf("providerSpec.volumes[%d] must be a valid UUID", i))
			}
		}
	}

	// Validate KeypairName
	if spec.KeypairName != "" {
		if len(spec.KeypairName) > 127 {
			errors = append(errors, fmt.Errorf("providerSpec.keypairName exceeds maximum length of 127 characters"))
		}
		if !keypairNameRegex.MatchString(spec.KeypairName) {
			errors = append(errors, fmt.Errorf("providerSpec.keypairName contains invalid characters (allowed: A-Za-z0-9@._-)"))
		}
	}

	// Validate AllowedAddresses
	if len(spec.AllowedAddresses) > 0 {
		for _, cidr := range spec.AllowedAddresses {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				errors = append(errors, fmt.Errorf("providerSpec.allowedAddresses has an invalid CIDR: %s", cidr))
			}
		}
	}

	// Validate AffinityGroup
	if spec.AffinityGroup != "" {
		if !isValidUUID(spec.AffinityGroup) {
			errors = append(errors, fmt.Errorf("providerSpec.affinityGroup must be a valid UUID"))
		}
	}

	// Validate ServiceAccountMails
	if len(spec.ServiceAccountMails) > 0 {
		// STACKIT API currently limits to 1 service account per server
		if len(spec.ServiceAccountMails) > 1 {
			errors = append(errors, fmt.Errorf("providerSpec.serviceAccountMails can contain a maximum of 1 service account (STACKIT API constraint)"))
		}
		for i, email := range spec.ServiceAccountMails {
			if email == "" {
				errors = append(errors, fmt.Errorf("providerSpec.serviceAccountMails[%d] cannot be empty", i))
			} else if !isValidEmail(email) {
				errors = append(errors, fmt.Errorf("providerSpec.serviceAccountMails[%d] must be a valid email address", i))
			}
		}
	}

	// Validate Agent
	// Agent is optional with no specific constraints - just a boolean flag
	// No validation needed as any value (nil, true, false) is acceptable

	// Validate Metadata
	// Metadata is optional with no specific constraints - freeform JSON object
	// No validation needed as any key-value pairs are acceptable

	return errors
}

// validateNetworking validates the NetworkingSpec
func validateNetworking(networking *api.NetworkingSpec) []error {
	var errors []error

	hasNetworkID := networking.NetworkID != ""
	hasNICIDs := len(networking.NICIDs) > 0

	// Either NetworkID or NICIDs must be set, but not both
	if !hasNetworkID && !hasNICIDs {
		errors = append(errors, fmt.Errorf("providerSpec.networking must specify either networkId or nicIds"))
		return errors
	}

	if hasNetworkID && hasNICIDs {
		errors = append(errors, fmt.Errorf("providerSpec.networking cannot specify both networkId and nicIds (mutually exclusive)"))
		return errors
	}

	// Validate NetworkID format if specified
	if hasNetworkID {
		if !isValidUUID(networking.NetworkID) {
			errors = append(errors, fmt.Errorf("providerSpec.networking.networkId must be a valid UUID"))
		}
	}

	// Validate NICIDs if specified
	if hasNICIDs {
		for i, nicID := range networking.NICIDs {
			if nicID == "" {
				errors = append(errors, fmt.Errorf("providerSpec.networking.nicIds[%d] cannot be empty", i))
			} else if !isValidUUID(nicID) {
				errors = append(errors, fmt.Errorf("providerSpec.networking.nicIds[%d] must be a valid UUID", i))
			}
		}
	}

	return errors
}

// validateBootVolume validates the BootVolumeSpec
func validateBootVolume(bootVolume *api.BootVolumeSpec) []error {
	var errors []error

	// Validate size if specified
	if bootVolume.Size < 0 {
		errors = append(errors, fmt.Errorf("providerSpec.bootVolume.size must be positive or zero"))
	}

	// Validate source if specified
	if bootVolume.Source != nil {
		if bootVolume.Source.Type == "" {
			errors = append(errors, fmt.Errorf("providerSpec.bootVolume.source.type is required when source is specified"))
		} else {
			// Validate source type is one of the allowed values
			validSourceTypes := map[string]bool{
				"image":    true,
				"snapshot": true,
				"volume":   true,
			}
			if !validSourceTypes[bootVolume.Source.Type] {
				errors = append(errors, fmt.Errorf("providerSpec.bootVolume.source type must be one of: image, snapshot, volume"))
			}
		}

		if bootVolume.Source.ID == "" {
			errors = append(errors, fmt.Errorf("providerSpec.bootVolume.source.id is required when source is specified"))
		} else if !isValidUUID(bootVolume.Source.ID) {
			errors = append(errors, fmt.Errorf("providerSpec.bootVolume.source.id must be a valid UUID"))
		}
	}

	return errors
}

// isValidUUID checks if a string is a valid UUID
func isValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// isValidEmail checks if a string is a valid email address
func isValidEmail(s string) bool {
	return emailRegex.MatchString(s)
}

// isValidMachineType checks if a string matches the machine type format
func isValidMachineType(s string) bool {
	return machineTypeRegex.MatchString(s)
}

// isValidRegion checks if a string matches the STACKIT region format
func isValidRegion(s string) bool {
	return regionRegex.MatchString(s)
}

// isValidAvailabilityZone checks if a string matches the STACKIT availability zone format
func isValidAvailabilityZone(s string) bool {
	return availabilityZoneRegex.MatchString(s)
}

// isValidJSON checks if a string is valid JSON
func isValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}
