package model

// region container labels
type LabelContainer string

var labelsContainer = []LabelContainer{
	ContainerID,
	ContainerIDFull,
	ContainerName,
	ContainerImageName,
	ContainerImageRevision,
	ContainerImageSource,
	ContainerImageTitle,
}

const (
	ContainerID     LabelContainer = "container_id"
	ContainerIDFull LabelContainer = "container_id_full"
	ContainerName   LabelContainer = "container_name"

	ContainerImageName     LabelContainer = "container_image_name"
	ContainerImageRevision LabelContainer = "container_image_revision"
	ContainerImageSource   LabelContainer = "container_image_source"
	ContainerImageTitle    LabelContainer = "container_image_title"
)

func (m *MetaLog) LabelsContainer() map[LabelContainer]string {
	result := make(map[LabelContainer]string)

	for _, label := range labelsContainer {
		if val, ok := m.Labels[string(label)]; ok {
			result[label] = val
		}
	}
	return result
}

//endregion

//region static labels

type LabelStatic string

var lablesStatic = []LabelStatic{
	StaticLabelIngress,
	StaticLabelJobType,
	StaticLabelStack,
	StaticLabelNameSpace,
	StaticLabelJob,
	StaticLabelTaskGroup,
	StaticLabelTask,
	StaticLabelHost,
	StaticLabelHostId,
}

const (
	StaticLabelIngress   LabelStatic = "ingress"
	StaticLabelJobType   LabelStatic = "job_type"
	StaticLabelStack     LabelStatic = "stack"
	StaticLabelNameSpace LabelStatic = "namespace"
	StaticLabelJob       LabelStatic = "job"
	StaticLabelTaskGroup LabelStatic = "task_group"
	StaticLabelTask      LabelStatic = "task"
	StaticLabelHost      LabelStatic = "host_name"
	StaticLabelHostId    LabelStatic = "host_id"
)

func (m *MetaLog) LabelsStatic() map[LabelStatic]string {
	result := make(map[LabelStatic]string)

	for _, label := range lablesStatic {
		if val, ok := m.Labels[string(label)]; ok {
			result[label] = val
		}
	}
	return result
}

//endregion

//region dynamic labels

// IndexedLabels expected in ecs.Labels
var labelsDynamic = []LabelDynamic{
	DynamicLabelLevel,
	DynamicLabelUsedGrok,
}

// LabelDynamic. Labels can only be emmited after anaylse the log message
type LabelDynamic string

const (
	DynamicLabelLevel LabelDynamic = "level"

	DynamicLabelUsedGrok LabelDynamic = "used_grok"
)

func (m *MetaLog) LabelsDynamic() map[LabelDynamic]string {
	result := make(map[LabelDynamic]string)

	for _, label := range labelsDynamic {
		if val, ok := m.Labels[string(label)]; ok {
			result[label] = val
		}
	}
	return result
}

//endregion

// region ecs meta field converters
func (m *MetaLog) EcsTags() []string {
	return m.Tags
}
func (m *MetaLog) EcsLabels() map[string]string {
	return m.Labels
}

func (m *MetaLog) EcsContainerInfo() *Container {
	containerLabels := m.LabelsContainer()
	if len(containerLabels) > 0 {
		return &Container{
			Id: containerLabels[ContainerID],
			Image: &Container_Image{
				Name: containerLabels[ContainerImageName],
				Tag:  nil,
			},
			Labels:    nil,
			Name:      containerLabels[ContainerName],
			Runtime:   "",
			CreatedAt: nil,
		}
	}
	return nil
}

func (m *MetaLog) EcsAgentInfo() *Agent {
	//return &Agent{
	//	Build:       nil,
	//	EphemeralId: "",
	//	Id:          "",
	//	Name:        "",
	//	Type:        "",
	//	Version:     "",
	//}
	return nil
}
func (m *MetaLog) EcsHostInfo() *Host {
	staticLabels := m.LabelsStatic()
	if len(staticLabels) > 0 {
		return &Host{
			Architecture: "",
			Domain:       "",
			Hostname:     staticLabels[StaticLabelHost],
			Id:           staticLabels[StaticLabelHostId],
			Ip:           nil,
			Mac:          nil,
			Name:         "",
			Os:           nil,
			Type:         "",
			Uptime:       0,
			User:         nil,
		}
	}
	return nil
}

func (m *MetaLog) EcsOrganizationInfo() *Organization {

	return nil
}

func (m *MetaLog) EcsServiceInfo() *Service {
	return nil
}

//endregion
