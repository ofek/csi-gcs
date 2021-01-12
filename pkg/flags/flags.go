package flags

import (
	"flag"
	"strconv"
	"strings"

	"k8s.io/klog"
)

const (
	FLAG_BUCKET              = "bucket"
	FLAG_PROJECT_ID          = "projectId"
	FLAG_KMS_KEY_ID          = "kmsKeyId"
	FLAG_LOCATION            = "location"
	FLAG_FUSE_MOUNT_OPTION   = "fuseMountOptions"
	FLAG_DIR_MODE            = "dirMode"
	FLAG_FILE_MODE           = "fileMode"
	FLAG_UID                 = "uid"
	FLAG_GID                 = "gid"
	FLAG_IMPLICIT_DIRS       = "implicitDirs"
	FLAG_BILLING_PROJECT     = "billingProject"
	FLAG_LIMIT_BYTES_PER_SEC = "limitBytesPerSec"
	FLAG_LIMIT_OPS_PER_SEC   = "limitOpsPerSec"
	FLAG_STAT_CACHE_TTL      = "statCacheTTL"
	FLAG_TYPE_CACHE_TTL      = "typeCacheTTL"

	ANNOTATION_PREFIX = "gcs.csi.ofek.dev/"

	ANNOTATION_BUCKET              = "gcs.csi.ofek.dev/bucket"
	ANNOTATION_PROJECT_ID          = "gcs.csi.ofek.dev/project-id"
	ANNOTATION_KMS_KEY_ID          = "gcs.csi.ofek.dev/kms-key-id"
	ANNOTATION_LOCATION            = "gcs.csi.ofek.dev/location"
	ANNOTATION_FUSE_MOUNT_OPTION   = "gcs.csi.ofek.dev/fuse-mount-options"
	ANNOTATION_DIR_MODE            = "gcs.csi.ofek.dev/dir-mode"
	ANNOTATION_FILE_MODE           = "gcs.csi.ofek.dev/file-mode"
	ANNOTATION_UID                 = "gcs.csi.ofek.dev/uid"
	ANNOTATION_GID                 = "gcs.csi.ofek.dev/gid"
	ANNOTATION_IMPLICIT_DIRS       = "gcs.csi.ofek.dev/implicit-dirs"
	ANNOTATION_BILLING_PROJECT     = "gcs.csi.ofek.dev/billing-project"
	ANNOTATION_LIMIT_BYTES_PER_SEC = "gcs.csi.ofek.dev/limit-bytes-per-sec"
	ANNOTATION_LIMIT_OPS_PER_SEC   = "gcs.csi.ofek.dev/limit-ops-per-sec"
	ANNOTATION_STAT_CACHE_TTL      = "gcs.csi.ofek.dev/stat-cache-ttl"
	ANNOTATION_TYPE_CACHE_TTL      = "gcs.csi.ofek.dev/type-cache-ttl"

	MOUNT_OPTION_BUCKET              = "bucket"
	MOUNT_OPTION_PROJECT_ID          = "project-id"
	MOUNT_OPTION_KMS_KEY_ID          = "kms-key-id"
	MOUNT_OPTION_LOCATION            = "location"
	MOUNT_OPTION_FUSE_MOUNT_OPTION   = "fuse-mount-option"
	MOUNT_OPTION_DIR_MODE            = "dir-mode"
	MOUNT_OPTION_FILE_MODE           = "file-mode"
	MOUNT_OPTION_UID                 = "uid"
	MOUNT_OPTION_GID                 = "gid"
	MOUNT_OPTION_IMPLICIT_DIRS       = "implicit-dirs"
	MOUNT_OPTION_BILLING_PROJECT     = "billing-project"
	MOUNT_OPTION_LIMIT_BYTES_PER_SEC = "limit-bytes-per-sec"
	MOUNT_OPTION_LIMIT_OPS_PER_SEC   = "limit-ops-per-sec"
	MOUNT_OPTION_STAT_CACHE_TTL      = "stat-cache-ttl"
	MOUNT_OPTION_TYPE_CACHE_TTL      = "type-cache-ttl"
)

func IsFlag(flag string) bool {
	switch flag {
	case FLAG_BUCKET:
		return true
	case FLAG_PROJECT_ID:
		return true
	case FLAG_KMS_KEY_ID:
		return true
	case FLAG_LOCATION:
		return true
	case FLAG_FUSE_MOUNT_OPTION:
		return true
	case FLAG_DIR_MODE:
		return true
	case FLAG_FILE_MODE:
		return true
	case FLAG_UID:
		return true
	case FLAG_GID:
		return true
	case FLAG_IMPLICIT_DIRS:
		return true
	case FLAG_BILLING_PROJECT:
		return true
	case FLAG_LIMIT_BYTES_PER_SEC:
		return true
	case FLAG_LIMIT_OPS_PER_SEC:
		return true
	case FLAG_STAT_CACHE_TTL:
		return true
	case FLAG_TYPE_CACHE_TTL:
		return true
	}
	return false
}

func FlagNameFromAnnotation(annotation string) string {
	switch annotation {
	case ANNOTATION_BUCKET:
		return FLAG_BUCKET
	case ANNOTATION_PROJECT_ID:
		return FLAG_PROJECT_ID
	case ANNOTATION_KMS_KEY_ID:
		return FLAG_KMS_KEY_ID
	case ANNOTATION_LOCATION:
		return FLAG_LOCATION
	case ANNOTATION_FUSE_MOUNT_OPTION:
		return FLAG_FUSE_MOUNT_OPTION
	case ANNOTATION_DIR_MODE:
		return FLAG_DIR_MODE
	case ANNOTATION_FILE_MODE:
		return FLAG_FILE_MODE
	case ANNOTATION_UID:
		return FLAG_UID
	case ANNOTATION_GID:
		return FLAG_GID
	case ANNOTATION_IMPLICIT_DIRS:
		return FLAG_IMPLICIT_DIRS
	case ANNOTATION_BILLING_PROJECT:
		return FLAG_BILLING_PROJECT
	case ANNOTATION_LIMIT_BYTES_PER_SEC:
		return FLAG_LIMIT_BYTES_PER_SEC
	case ANNOTATION_LIMIT_OPS_PER_SEC:
		return FLAG_LIMIT_OPS_PER_SEC
	case ANNOTATION_STAT_CACHE_TTL:
		return FLAG_STAT_CACHE_TTL
	case ANNOTATION_TYPE_CACHE_TTL:
		return FLAG_TYPE_CACHE_TTL
	}
	return ""
}

func IsOwnAnnotation(annotation string) bool {
	return strings.HasPrefix(annotation, ANNOTATION_PREFIX)
}

func IsAnnotation(annotation string) bool {
	return FlagNameFromAnnotation(annotation) != ""
}

func FlagNameFromMountOption(cmd string) string {
	switch cmd {
	case MOUNT_OPTION_BUCKET:
		return FLAG_BUCKET
	case MOUNT_OPTION_PROJECT_ID:
		return FLAG_PROJECT_ID
	case MOUNT_OPTION_KMS_KEY_ID:
		return FLAG_KMS_KEY_ID
	case MOUNT_OPTION_LOCATION:
		return FLAG_LOCATION
	case MOUNT_OPTION_FUSE_MOUNT_OPTION:
		return FLAG_FUSE_MOUNT_OPTION
	case MOUNT_OPTION_DIR_MODE:
		return FLAG_DIR_MODE
	case MOUNT_OPTION_FILE_MODE:
		return FLAG_FILE_MODE
	case MOUNT_OPTION_UID:
		return FLAG_UID
	case MOUNT_OPTION_GID:
		return FLAG_GID
	case MOUNT_OPTION_IMPLICIT_DIRS:
		return FLAG_IMPLICIT_DIRS
	case MOUNT_OPTION_BILLING_PROJECT:
		return FLAG_BILLING_PROJECT
	case MOUNT_OPTION_LIMIT_BYTES_PER_SEC:
		return FLAG_LIMIT_BYTES_PER_SEC
	case MOUNT_OPTION_LIMIT_OPS_PER_SEC:
		return FLAG_LIMIT_OPS_PER_SEC
	case MOUNT_OPTION_STAT_CACHE_TTL:
		return FLAG_STAT_CACHE_TTL
	case MOUNT_OPTION_TYPE_CACHE_TTL:
		return FLAG_TYPE_CACHE_TTL
	}
	return ""
}

func IsMountOption(cmd string) bool {
	return FlagNameFromMountOption(cmd) != ""
}

func MergeFlags(a map[string]string, b map[string]string) (result map[string]string) {
	result = a

	for k, v := range b {
		if !IsFlag(k) {
			klog.Warningf("Flag %s unknown", k)
			continue
		}
		result[k] = v
	}

	return result
}

func MergeSecret(a map[string]string, b map[string]string) (result map[string]string) {
	result = a

	for k, v := range b {
		if !IsFlag(k) {
			continue
		}
		result[k] = v
	}

	return result
}

func MergeAnnotations(a map[string]string, b map[string]string) (result map[string]string) {
	result = a

	for k, v := range b {
		if !IsOwnAnnotation(k) {
			continue
		}
		if !IsAnnotation(k) {
			klog.Warningf("Annotation %s unknown", k)
			continue
		}
		result[FlagNameFromAnnotation(k)] = v
	}

	return result
}

type fuseMountOptions []string

func (i *fuseMountOptions) String() string {
	return ""
}

func (i *fuseMountOptions) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type octalInt int64

func (i *octalInt) String() string {
	return ""
}

func (i *octalInt) Set(value string) error {
	parsedInt, err := strconv.ParseInt(value, 8, 64)

	*i = octalInt(parsedInt)

	return err
}

func MergeMountOptions(a map[string]string, b []string) (result map[string]string) {
	var (
		args             = flag.NewFlagSet("csi-gcs", flag.ContinueOnError)
		bucket           string
		projectId        string
		kmsKeyId         string
		location         string
		fuseMountOptions fuseMountOptions
		dirMode          octalInt = -1
		fileMode         octalInt = -1
		uid              int64
		gid              int64
		implicitDirs     bool
		billingProject   string
		limitBytesPerSec int64
		limitOpsPerSec   int64
		statCacheTTL     string
		typeCacheTTL     string
	)

	args.StringVar(&bucket, MOUNT_OPTION_BUCKET, "", "Bucket Name")
	args.StringVar(&projectId, MOUNT_OPTION_PROJECT_ID, "", "Project ID of Bucket")
	args.StringVar(&kmsKeyId, MOUNT_OPTION_KMS_KEY_ID, "", "KMS encryption key ID. (projects/my-pet-project/locations/us-east1/keyRings/my-key-ring/cryptoKeys/my-key)")
	args.StringVar(&location, MOUNT_OPTION_LOCATION, "", "Bucket Location")
	args.Var(&fuseMountOptions, MOUNT_OPTION_FUSE_MOUNT_OPTION, "")
	args.Var(&dirMode, MOUNT_OPTION_DIR_MODE, "Permission bits for directories, in octal. (default: 0775)")
	args.Var(&fileMode, MOUNT_OPTION_FILE_MODE, "Permission bits for files, in octal. (default: 0664)")
	args.Int64Var(&uid, MOUNT_OPTION_UID, -1, "UID owner of all inodes. (default: -1)")
	args.Int64Var(&gid, MOUNT_OPTION_GID, -1, "GID owner of all inodes. (default: -1)")
	args.BoolVar(&implicitDirs, MOUNT_OPTION_IMPLICIT_DIRS, false, "Implicitly define directories based on content.")
	args.StringVar(&billingProject, MOUNT_OPTION_BILLING_PROJECT, "", "Project to use for billing when accessing requester pays buckets.")
	args.Int64Var(&limitBytesPerSec, MOUNT_OPTION_LIMIT_BYTES_PER_SEC, -1, "Bandwidth limit for reading data, measured over a 30-second window.")
	args.Int64Var(&limitOpsPerSec, MOUNT_OPTION_LIMIT_OPS_PER_SEC, -1, "Operations per second limit, measured over a 30-second window.")
	args.StringVar(&statCacheTTL, MOUNT_OPTION_STAT_CACHE_TTL, "", "How long to cache StatObject results and inode attributes.")
	args.StringVar(&typeCacheTTL, MOUNT_OPTION_TYPE_CACHE_TTL, "", "How long to cache name -> file/dir mappings in directory inodes.")

	err := args.Parse(b)
	if err != nil {
		klog.Warningf("%s", err)
	}

	result = a

	if bucket != "" {
		result[FLAG_BUCKET] = bucket
	}

	if projectId != "" {
		result[FLAG_PROJECT_ID] = projectId
	}

	if kmsKeyId != "" {
		result[FLAG_KMS_KEY_ID] = kmsKeyId
	}

	if location != "" {
		result[FLAG_LOCATION] = location
	}

	if len(fuseMountOptions) != 0 {
		result[FLAG_FUSE_MOUNT_OPTION] = strings.Join(fuseMountOptions, ",")
	}

	if dirMode != -1 {
		result[FLAG_DIR_MODE] = "0" + strconv.FormatInt(int64(dirMode), 8)
	}

	if fileMode != -1 {
		result[FLAG_FILE_MODE] = "0" + strconv.FormatInt(int64(fileMode), 8)
	}

	if uid != -1 {
		result[FLAG_UID] = strconv.FormatInt(uid, 10)
	}

	if gid != -1 {
		result[FLAG_GID] = strconv.FormatInt(gid, 10)
	}

	if implicitDirs {
		result[FLAG_IMPLICIT_DIRS] = "true"
	}

	if billingProject != "" {
		result[FLAG_BILLING_PROJECT] = billingProject
	}

	if limitBytesPerSec != -1 {
		result[FLAG_LIMIT_BYTES_PER_SEC] = strconv.FormatInt(limitBytesPerSec, 10)
	}

	if limitOpsPerSec != -1 {
		result[FLAG_LIMIT_OPS_PER_SEC] = strconv.FormatInt(limitOpsPerSec, 10)
	}

	if statCacheTTL != "" {
		result[FLAG_STAT_CACHE_TTL] = statCacheTTL
	}

	if typeCacheTTL != "" {
		result[FLAG_TYPE_CACHE_TTL] = typeCacheTTL
	}

	return result
}

func FlagNameToGcsfuseOption(flag string) string {
	switch flag {
	case FLAG_DIR_MODE:
		return "dir_mode"
	case FLAG_FILE_MODE:
		return "file_mode"
	case FLAG_UID:
		return "uid"
	case FLAG_GID:
		return "gid"
	case FLAG_IMPLICIT_DIRS:
		return "implicit_dirs"
	case FLAG_BILLING_PROJECT:
		return "billing_project"
	case FLAG_LIMIT_BYTES_PER_SEC:
		return "limit_bytes_per_sec"
	case FLAG_LIMIT_OPS_PER_SEC:
		return "limit_ops_per_sec"
	case FLAG_STAT_CACHE_TTL:
		return "stat_cache_ttl"
	case FLAG_TYPE_CACHE_TTL:
		return "type_cache_ttl"
	}
	return ""
}

func MaybeAddFlag(result []string, flags map[string]string, name string) []string {
	argName := FlagNameToGcsfuseOption(name)

	value, found := flags[name]
	if found {
		result = append(result, argName+"="+value)
	}
	return result
}
func MaybeAddBooleanFlag(result []string, flags map[string]string, name string) []string {
	argName := FlagNameToGcsfuseOption(name)

	value, found := flags[name]
	if found && value == "true" {
		result = append(result, argName)
	}
	return result
}

func MaybeAddDirectFlag(result []string, flags map[string]string, name string) []string {
	value, found := flags[name]
	if found {
		result = append(result, strings.Split(value, ",")...)
	}
	return result
}

func ExtraFlags(flags map[string]string) (result []string) {
	result = []string{}

	result = MaybeAddDirectFlag(result, flags, FLAG_FUSE_MOUNT_OPTION)
	result = MaybeAddFlag(result, flags, FLAG_DIR_MODE)
	result = MaybeAddFlag(result, flags, FLAG_FILE_MODE)
	result = MaybeAddFlag(result, flags, FLAG_UID)
	result = MaybeAddFlag(result, flags, FLAG_GID)
	result = MaybeAddBooleanFlag(result, flags, FLAG_IMPLICIT_DIRS)
	result = MaybeAddFlag(result, flags, FLAG_LIMIT_BYTES_PER_SEC)
	result = MaybeAddFlag(result, flags, FLAG_LIMIT_OPS_PER_SEC)
	result = MaybeAddFlag(result, flags, FLAG_STAT_CACHE_TTL)
	result = MaybeAddFlag(result, flags, FLAG_TYPE_CACHE_TTL)

	return result
}
