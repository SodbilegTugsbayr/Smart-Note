export default defineNuxtPlugin((NuxtApp) => {
  return {
    provide: {
      fullDate: function (dateStr: string | Date, separator: string) {
        return fullDate(new Date(dateStr), separator)
      },
      coolDate: function (dateStr: string | Date) {
        return coolDate(new Date(dateStr))
      },
      coolDate2: function (dateStr: string | Date) {
        return coolDate(new Date(dateStr))
      },
      onlyDate: function (dateStr: string | Date) {
        return onlyDate(new Date(dateStr))
      },
      onlyTime: function (dateStr: string | Date) {
        return onlyTime(new Date(dateStr))
      },
      formatDuration: function (sec_num: number) {
        return formatDuration(sec_num)
      },
      formatSecond: function (sec_num: number) {
        return formatSecond(sec_num)
      },
      humanSize,
      formatMoney,
    },
  }
})

const GB = 1000 * 1000 * 1000,
  MB = 1000 * 1000,
  KB = 1000

const humanSize = function (size: number): string {
  if (size > GB) {
    return (size / GB).toFixed(1) + " GB"
  } else if (size > MB) {
    return (size / MB).toFixed(1) + " MB"
  } else if (size > KB) {
    return (size / KB).toFixed(1) + " KB"
  } else {
    return size + " B"
  }
}

function formatMoney(number: number, lang: string = "mn-Mn", minFractionDigit: number = 0, maxFractionDigit: number = 0): string {
  let formatted = ""
  if (isNaN(number)) {
    formatted = ""
  } else {
    formatted = Number(number).toLocaleString(lang, {
      maximumFractionDigits: maxFractionDigit,
      minimumFractionDigits: minFractionDigit,
    })
  }
  return formatted
}

const formatDuration = function (sec_num: number) {
  let hours = Math.floor(sec_num / 3600)
  let minutes = Math.floor((sec_num - hours * 3600) / 60)
  let seconds = Math.round(sec_num - hours * 3600 - minutes * 60)

  let hoursStr = hours.toString()
  let minutesStr = minutes.toString()
  let secondsStr = seconds.toString()
  if (hours < 10) {
    hoursStr = "0" + hours
  }
  if (minutes < 10) {
    minutesStr = "0" + minutes
  }
  if (seconds < 10) {
    secondsStr = "0" + seconds
  }
  return hoursStr + ":" + minutesStr + ":" + secondsStr
}

const formatSecond = function (sec_num: number) {
  let minutes = Math.floor(sec_num / 60)
  let seconds = Math.round(sec_num - minutes * 60)

  let minutesStr = minutes.toString()
  let secondsStr = seconds.toString()
  if (minutes < 10) {
    minutesStr = "0" + minutes
  }
  if (seconds < 10) {
    secondsStr = "0" + seconds
  }
  return minutesStr + ":" + secondsStr
}

const fullDate = function (date: Date, separator: string = '/') {
  return onlyDate(date).replaceAll("-", separator) + ` ${onlyTime(date)}`
}

const today = new Date()

const dateDiffInDays = (a: Date, b: Date) => {
  const MS_PER_DAY = 1000 * 60 * 60 * 24
  const utc1 = Date.UTC(a.getFullYear(), a.getMonth(), a.getDate())
  const utc2 = Date.UTC(b.getFullYear(), b.getMonth(), b.getDate())

  return Math.floor((utc2 - utc1) / MS_PER_DAY)
}

const relativeDate = (date: Date) => {
  const diffDays = dateDiffInDays(today, date)
  let relativeDateStr = ""
  if (diffDays == 0) relativeDateStr += "Өнөөдөр"
  if (diffDays == 1) relativeDateStr += "Маргааш"
  if (diffDays == 2) relativeDateStr += "Нөгөөдөр"
  if (diffDays == -1) relativeDateStr += "Өчигдөр"
  if (diffDays == -2) relativeDateStr += "Уржигдар"
  if (diffDays > 2 || diffDays < -2) relativeDateStr += onlyDate(date).replaceAll("-", "/")

  return (relativeDateStr += ` ${onlyTime(date)}`)
}

const coolDate = function (date: Date) {
  const seconds = Math.abs(Math.floor((new Date().valueOf() - date.valueOf()) / 1000))

  const isPast = new Date() > date
  if (seconds < 86400) {
    let interval = Math.floor(seconds / 3600)
    if (interval >= 1) {
      if (interval == 1) {
        return "1 " + `цагийн ${isPast ? "өмнө" : "дараа"}`
      }
      return interval + " " + `цагийн ${isPast ? "өмнө" : "дараа"}`
    }
    interval = Math.floor(seconds / 60)
    if (interval >= 1) {
      if (interval == 1) {
        return "1 " + `минутын ${isPast ? "өмнө" : "дараа"}`
      }
      return interval + " " + `минутын ${isPast ? "өмнө" : "дараа"}`
    }
    return "Одоо"
  } else {
    return relativeDate(date)
  }
}

const coolDate2 = function (dateTime: Date) {
  let nowTime = new Date();

  const secondsPast = Math.floor((nowTime.getTime() - dateTime.getTime()) / 1000);

  if (secondsPast < 60) {
    return secondsPast + " секундийн өмнө";
  }
  if (secondsPast < 3600) {
    return Math.floor(secondsPast / 60) + " минутын өмнө";
  }
  if (secondsPast < 86400) {
    return Math.floor(secondsPast / 3600) + " цагийн өмнө";
  }
  if (secondsPast < 2592000) { // 30 days
    return Math.floor(secondsPast / 86400) + " өдрийн өмнө";
  }
  if (secondsPast < 31104000) { // 12 months
    return Math.floor(secondsPast / 2592000) + " сарын өмнө";
  }
  return Math.floor(secondsPast / 31104000) + " жилийн өмнө";
}

const onlyDate = function (date: Date) {
  return (
    date.getFullYear() +
    "-" +
    ("0" + (date.getMonth() + 1)).slice(-2) +
    "-" +
    ("0" + date.getDate()).slice(-2)
  )
}

const onlyTime = function (date: Date) {
  let dateTime = new Date(date)
  let hour = dateTime.getHours()
  let minute = dateTime.getMinutes()
  let h = hour < 10 ? "0" + hour : hour.toString(),
    m = minute < 10 ? "0" + minute : minute.toString()
  return h + `:` + m
}
