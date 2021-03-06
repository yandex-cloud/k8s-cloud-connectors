// Code generated by protoc-gen-goext. DO NOT EDIT.

package ydb

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

type Database_DatabaseType = isDatabase_DatabaseType

func (m *Database) SetDatabaseType(v Database_DatabaseType) {
	m.DatabaseType = v
}

func (m *Database) SetId(v string) {
	m.Id = v
}

func (m *Database) SetFolderId(v string) {
	m.FolderId = v
}

func (m *Database) SetCreatedAt(v *timestamp.Timestamp) {
	m.CreatedAt = v
}

func (m *Database) SetName(v string) {
	m.Name = v
}

func (m *Database) SetDescription(v string) {
	m.Description = v
}

func (m *Database) SetStatus(v Database_Status) {
	m.Status = v
}

func (m *Database) SetEndpoint(v string) {
	m.Endpoint = v
}

func (m *Database) SetResourcePresetId(v string) {
	m.ResourcePresetId = v
}

func (m *Database) SetStorageConfig(v *StorageConfig) {
	m.StorageConfig = v
}

func (m *Database) SetScalePolicy(v *ScalePolicy) {
	m.ScalePolicy = v
}

func (m *Database) SetNetworkId(v string) {
	m.NetworkId = v
}

func (m *Database) SetSubnetIds(v []string) {
	m.SubnetIds = v
}

func (m *Database) SetZonalDatabase(v *ZonalDatabase) {
	m.DatabaseType = &Database_ZonalDatabase{
		ZonalDatabase: v,
	}
}

func (m *Database) SetRegionalDatabase(v *RegionalDatabase) {
	m.DatabaseType = &Database_RegionalDatabase{
		RegionalDatabase: v,
	}
}

func (m *Database) SetDedicatedDatabase(v *DedicatedDatabase) {
	m.DatabaseType = &Database_DedicatedDatabase{
		DedicatedDatabase: v,
	}
}

func (m *Database) SetServerlessDatabase(v *ServerlessDatabase) {
	m.DatabaseType = &Database_ServerlessDatabase{
		ServerlessDatabase: v,
	}
}

func (m *Database) SetAssignPublicIps(v bool) {
	m.AssignPublicIps = v
}

func (m *Database) SetLocationId(v string) {
	m.LocationId = v
}

func (m *Database) SetLabels(v map[string]string) {
	m.Labels = v
}

func (m *Database) SetBackupConfig(v *BackupConfig) {
	m.BackupConfig = v
}

func (m *Database) SetDocumentApiEndpoint(v string) {
	m.DocumentApiEndpoint = v
}

func (m *DedicatedDatabase) SetResourcePresetId(v string) {
	m.ResourcePresetId = v
}

func (m *DedicatedDatabase) SetStorageConfig(v *StorageConfig) {
	m.StorageConfig = v
}

func (m *DedicatedDatabase) SetScalePolicy(v *ScalePolicy) {
	m.ScalePolicy = v
}

func (m *DedicatedDatabase) SetNetworkId(v string) {
	m.NetworkId = v
}

func (m *DedicatedDatabase) SetSubnetIds(v []string) {
	m.SubnetIds = v
}

func (m *DedicatedDatabase) SetAssignPublicIps(v bool) {
	m.AssignPublicIps = v
}

func (m *ZonalDatabase) SetZoneId(v string) {
	m.ZoneId = v
}

func (m *RegionalDatabase) SetRegionId(v string) {
	m.RegionId = v
}

type ScalePolicy_ScaleType = isScalePolicy_ScaleType

func (m *ScalePolicy) SetScaleType(v ScalePolicy_ScaleType) {
	m.ScaleType = v
}

func (m *ScalePolicy) SetFixedScale(v *ScalePolicy_FixedScale) {
	m.ScaleType = &ScalePolicy_FixedScale_{
		FixedScale: v,
	}
}

func (m *ScalePolicy_FixedScale) SetSize(v int64) {
	m.Size = v
}

func (m *StorageConfig) SetStorageOptions(v []*StorageOption) {
	m.StorageOptions = v
}

func (m *StorageOption) SetStorageTypeId(v string) {
	m.StorageTypeId = v
}

func (m *StorageOption) SetGroupCount(v int64) {
	m.GroupCount = v
}
