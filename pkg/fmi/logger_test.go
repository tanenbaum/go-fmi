package fmi

import (
	"errors"
	"reflect"
	"testing"
)

func Test_loggerCategory_String(t *testing.T) {
	tests := []struct {
		name string
		l    loggerCategory
		want string
	}{
		{
			"unknown",
			0,
			"unknown",
		},
		{
			"events",
			loggerCategoryEvents,
			"logEvents",
		},
		{
			"warning",
			loggerCategoryWarning,
			"logStatusWarning",
		},
		{
			"discard",
			loggerCategoryDiscard,
			"logStatusDiscard",
		},
		{
			"error",
			loggerCategoryError,
			"logStatusError",
		},
		{
			"fatal",
			loggerCategoryFatal,
			"logStatusFatal",
		},
		{
			"pending",
			loggerCategoryPending,
			"logStatusPending",
		},
		{
			"all",
			loggerCategoryAll,
			"logAll",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.String(); got != tt.want {
				t.Errorf("loggerCategory.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newLogger(t *testing.T) {
	type args struct {
		categories []string
		callback   loggerCallback
	}
	tests := []struct {
		name    string
		args    args
		want    Logger
		wantErr bool
	}{
		{
			"error returned with invalid logger category",
			args{
				[]string{"foo"},
				nil,
			},
			nil,
			true,
		},
		{
			"categories form bitmask",
			args{
				[]string{"logEvents", "logStatusError"},
				nil,
			},
			logger{
				mask: loggerCategoryEvents | loggerCategoryError,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newLogger(tt.args.categories, tt.args.callback)
			if (err != nil) != tt.wantErr {
				t.Errorf("newLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockLogger struct {
	status   Status
	category string
	message  string
}

func (m *mockLogger) callback(status Status, category, message string) {
	m.status = status
	m.category = category
	m.message = message
}

func Test_logger_Error(t *testing.T) {
	type fields struct {
		mask   loggerCategory
		logger *mockLogger
	}
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *mockLogger
	}{
		{
			"error callback is logged",
			fields{
				loggerCategoryError,
				&mockLogger{},
			},
			args{
				errors.New("foo"),
			},
			&mockLogger{
				status:   StatusError,
				category: "logStatusError",
				message:  "foo",
			},
		},
		{
			"error callback is not logged if mask is incorrect",
			fields{
				loggerCategoryWarning,
				&mockLogger{},
			},
			args{
				errors.New("foo"),
			},
			&mockLogger{},
		},
		{
			"error callback is logged if all is set",
			fields{
				loggerCategoryAll,
				&mockLogger{},
			},
			args{
				errors.New("foo"),
			},
			&mockLogger{
				status:   StatusError,
				category: "logStatusError",
				message:  "foo",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logger{
				mask:              tt.fields.mask,
				fmiCallbackLogger: tt.fields.logger.callback,
			}
			l.Error(tt.args.err)
			if !reflect.DeepEqual(tt.fields.logger, tt.want) {
				t.Errorf("Expect logger %v got %v", tt.want, tt.fields.logger)
			}
		})
	}
}
