package ringbuffer

import "testing"

func TestNew(t *testing.T) {
	rb := New[int](4)

	if rb.Capacity() != 4 {
		t.Fatalf("expected capacity 4")
	}

	if !rb.IsEmpty() {
		t.Fatal("new buffer should be empty")
	}
}

func TestEnqueueDequeue(t *testing.T) {
	rb := New[int](2)

	if err := rb.Enqueue(10); err != nil {
		t.Fatal(err)
	}

	if err := rb.Enqueue(20); err != nil {
		t.Fatal(err)
	}

	v, err := rb.Dequeue()
	if err != nil {
		t.Fatal(err)
	}

	if v != 10 {
		t.Fatalf("expected 10, got %d", v)
	}

	v, err = rb.Dequeue()
	if err != nil {
		t.Fatal(err)
	}

	if v != 20 {
		t.Fatalf("expected 20, got %d", v)
	}

	if !rb.IsEmpty() {
		t.Fatal("buffer should be empty")
	}
}

func TestBufferFull(t *testing.T) {
	rb := New[int](1)

	_ = rb.Enqueue(1)

	if err := rb.Enqueue(2); err != ErrBufferFull {
		t.Fatal("expected ErrBufferFull")
	}
}

func TestBufferEmpty(t *testing.T) {
	rb := New[int](1)

	_, err := rb.Dequeue()

	if err != ErrBufferEmpty {
		t.Fatal("expected ErrBufferEmpty")
	}
}

func TestWrapAround(t *testing.T) {
	rb := New[int](3)

	_ = rb.Enqueue(1)
	_ = rb.Enqueue(2)
	_ = rb.Enqueue(3)

	v, _ := rb.Dequeue()
	if v != 1 {
		t.Fatal("expected 1")
	}

	_ = rb.Enqueue(4)

	values := []int{}

	for !rb.IsEmpty() {
		v, _ := rb.Dequeue()
		values = append(values, v)
	}

	expected := []int{2, 3, 4}

	for i := range expected {
		if values[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, values)
		}
	}
}