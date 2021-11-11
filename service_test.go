package gridon

import (
	"reflect"
	"testing"
)

func Test_service_runnableContractTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		service             *service
		want1               bool
		wantContractRunning bool
	}{
		{name: "実行中ならfalse",
			service:             &service{contractRunning: true},
			want1:               false,
			wantContractRunning: true},
		{name: "実行中でないなら実行中に変更してtrue",
			service:             &service{contractRunning: false},
			want1:               true,
			wantContractRunning: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.service.runnableContractTask()
			if !reflect.DeepEqual(test.want1, got1) || !reflect.DeepEqual(test.wantContractRunning, test.service.contractRunning) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantContractRunning, got1, test.service.contractRunning)
			}
		})
	}
}

func Test_service_finishContractTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		service             *service
		wantContractRunning bool
	}{
		{name: "実行中ならfalseにできる",
			service:             &service{contractRunning: true},
			wantContractRunning: false},
		{name: "実行中でなくてもfalseにできる",
			service:             &service{contractRunning: false},
			wantContractRunning: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.service.finishContractTask()
			if !reflect.DeepEqual(test.wantContractRunning, test.service.contractRunning) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantContractRunning, test.service.contractRunning)
			}
		})
	}
}

func Test_service_runnableOrderTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		service          *service
		want1            bool
		wantOrderRunning bool
	}{
		{name: "実行中ならfalse",
			service:          &service{orderRunning: true},
			want1:            false,
			wantOrderRunning: true},
		{name: "実行中でないなら実行中に変更してtrue",
			service:          &service{orderRunning: false},
			want1:            true,
			wantOrderRunning: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1 := test.service.runnableOrderTask()
			if !reflect.DeepEqual(test.want1, got1) || !reflect.DeepEqual(test.wantOrderRunning, test.service.orderRunning) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.wantOrderRunning, got1, test.service.orderRunning)
			}
		})
	}
}

func Test_service_finishOrderTask(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		service          *service
		wantOrderRunning bool
	}{
		{name: "実行中ならfalseにできる",
			service:          &service{orderRunning: true},
			wantOrderRunning: false},
		{name: "実行中でなくてもfalseにできる",
			service:          &service{orderRunning: false},
			wantOrderRunning: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.service.finishOrderTask()
			if !reflect.DeepEqual(test.wantOrderRunning, test.service.orderRunning) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantOrderRunning, test.service.orderRunning)
			}
		})
	}
}
