/** Định dạng thời gian hiển thị theo giờ Việt Nam (DB lưu UTC) */
const VN = new Intl.DateTimeFormat('vi-VN', {
  timeZone: 'Asia/Ho_Chi_Minh',
  dateStyle: 'short',
  timeStyle: 'short',
})

export function formatVN(iso: string | null | undefined): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return VN.format(d)
}

/** "trong 5 phút" / "3 giờ trước" — mô tả tương đối */
export function fromNow(iso: string | null | undefined): string {
  if (!iso) return '—'
  const diffMs = Date.parse(iso) - Date.now()
  const abs = Math.abs(diffMs)
  const phut = Math.round(abs / 60000)
  const gio = Math.round(phut / 60)
  const ngay = Math.round(gio / 24)

  const donVi = ngay >= 1 ? `${ngay} ngày` : gio >= 1 ? `${gio} giờ` : `${Math.max(phut, 1)} phút`
  return diffMs >= 0 ? `trong ${donVi}` : `${donVi} trước`
}

/** ISO local cho <input type="datetime-local"> (giờ VN) */
export function toLocalInput(iso: string | null | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const vn = new Date(d.getTime() + 7 * 3600 * 1000)
  return vn.toISOString().slice(0, 16)
}

/** Ngược lại: từ <input type="datetime-local"> (giờ VN) sang ISO UTC */
export function fromLocalInput(value: string): string {
  if (!value) return ''
  return new Date(`${value}:00.000+07:00`).toISOString()
}
