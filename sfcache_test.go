package sfcache

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_sfcache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	key := "testKey"
	value := 1

	group := New[string, int](1, nil, 0)

	// First call
	val, _, err := group.Do(ctx, key, true, func(_ context.Context) (int, error) {
		return value, nil
	})
	require.NoError(t, err)
	require.Equal(t, value, val)

	// call using cache
	var called bool
	val, _, err = group.Do(ctx, key, true, func(_ context.Context) (int, error) {
		called = true
		return value, nil
	})
	require.NoError(t, err)
	require.Equal(t, value, val)
	require.False(t, called)

	// call without using cache
	val, _, err = group.Do(ctx, key, false, func(_ context.Context) (int, error) {
		called = true
		return value, nil
	})
	require.NoError(t, err)
	require.Equal(t, value, val)
	require.True(t, called)

	// clear cache
	group.Clear()
	called = false
	val, _, err = group.Do(ctx, key, true, func(_ context.Context) (int, error) {
		called = true
		return value, nil
	})
	require.NoError(t, err)
	require.Equal(t, value, val)
	require.True(t, called)

	// error
	errTest := errors.New("error")
	val, _, err = group.Do(ctx, key, false, func(_ context.Context) (int, error) {
		return 0, errTest
	})
	require.Equal(t, errTest, err)
	require.Equal(t, 0, val)
}
