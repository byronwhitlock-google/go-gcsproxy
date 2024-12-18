package main

import (
	"context"
	"fmt"
	"io"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/api/label"
	"google.golang.org/genproto/googleapis/api/metric"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredres "google.golang.org/genproto/googleapis/api/monitoredres"
)

// [START monitoring_write_timeseries]

// writeTimeSeriesValue writes a value for the custom metric created
func writeTimeSeriesValue(projectID, metricType string, latency float64, mode string, request_id string) error {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close()
	now := &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + projectID,
		TimeSeries: []*monitoringpb.TimeSeries{{
			Metric: &metricpb.Metric{
				Type: metricType,
				Labels: map[string]string{
					"mode":       mode,
					"request_id": request_id,
				},
			},
			Resource: &monitoredres.MonitoredResource{
				Type: "k8s_container",
				Labels: map[string]string{
					"location":       "us-central1",
					"cluster_name":   "gcs-proxy-poc",
					"namespace_name": "gcs-proxy-go",
					"pod_name":       "gcsproxy-podname",
					"container_name": "mitmproxy",
				},
			},
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					StartTime: now,
					EndTime:   now,
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_DoubleValue{
						DoubleValue: latency,
					},
				},
			}},
		}},
	}
	fmt.Println("writeTimeseriesRequest: ", req)

	err = c.CreateTimeSeries(ctx, req)
	if err != nil {
		return fmt.Errorf("could not write time series value, %w ", err)
	}
	return nil
}

// [END monitoring_write_timeseries]

// createCustomMetric creates a custom metric specified by the metric type.
func createCustomMetric(w io.Writer, projectID, metricType string) (*metricpb.MetricDescriptor, error) {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	md := &metric.MetricDescriptor{
		Name: "GCS Proxy Latency",
		Type: metricType,
		Labels: []*label.LabelDescriptor{{
			Key:         "mode",
			ValueType:   label.LabelDescriptor_STRING,
			Description: "GCS Proxy Mode. This value must be encryption or decryption.",
		}, {
			Key:         "request_id",
			ValueType:   label.LabelDescriptor_STRING,
			Description: "GCS request id's.",
		}},
		MetricKind:  metric.MetricDescriptor_GAUGE,
		ValueType:   metric.MetricDescriptor_DOUBLE,
		Unit:        "s",
		Description: "Custom Metric for encryption/decryption lattency of GCS go proxy.",
		DisplayName: "GCS Proxy Latency",
	}
	req := &monitoringpb.CreateMetricDescriptorRequest{
		Name:             "projects/" + projectID,
		MetricDescriptor: md,
	}
	m, err := c.CreateMetricDescriptor(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not create custom metric: %w", err)
	}
	fmt.Println(w, "Created ", m.GetName())

	return m, nil
}

// [END monitoring_create_metric]
