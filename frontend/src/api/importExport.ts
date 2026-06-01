import request from '@/utils/request'

export const exportTransactions = (startDate?: string, endDate?: string) => {
  const params: any = {}
  if (startDate) params.start_date = startDate
  if (endDate) params.end_date = endDate

  return request.get('/io/export', {
    params,
    responseType: 'blob'
  })
}

export const importTransactions = (file: File): Promise<any> => {
  const formData = new FormData()
  formData.append('file', file)

  return request.post('/io/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}
