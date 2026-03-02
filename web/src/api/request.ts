import axios from 'axios'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

request.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const msg = err.response?.data?.error || err.message
    return Promise.reject(new Error(msg))
  }
)

export default request
