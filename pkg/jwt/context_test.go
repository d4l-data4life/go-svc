package jwt

import (
	"context"
	"reflect"
	"testing"
)

func TestContext(t *testing.T) {
	t.Run("retrieves a saved map", func(t *testing.T) {
		m1 := Claims{
			Issuer: IssuerGesundheitscloud,
		}

		ctx := NewContext(context.Background(), &m1)

		m2, ok := fromContext(ctx)

		if !ok {
			t.Errorf("expected fromContext to return true when context carries the value")
		}

		if !reflect.DeepEqual(&m1, m2) {
			t.Errorf("expected: %+v\tfound:%+v", &m1, m2)
		}
	})
}
