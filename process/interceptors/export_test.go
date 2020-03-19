package interceptors

func (mdi *MultiDataInterceptor) Topic() string {
	return mdi.topic
}

func (sdi *SingleDataInterceptor) Topic() string {
	return sdi.topic
}
