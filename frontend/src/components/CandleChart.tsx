import { useEffect, useRef } from 'react'
import { createChart } from 'lightweight-charts'
import type { IChartApi, ISeriesApi, CandlestickData, Time } from 'lightweight-charts'

interface CandleChartProps {
  data: Array<{
    timestamp: string
    open: number
    high: number
    low: number
    close: number
    volume?: number
  }>
  height?: number
  lastPrice?: number
}

export default function CandleChart({ data, height = 400, lastPrice }: CandleChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candleSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const lastDataRef = useRef<typeof data>([])

  useEffect(() => {
    if (!chartContainerRef.current) return

    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { color: '#ffffff' },
        textColor: '#333',
      },
      grid: {
        vertLines: { color: '#f0f0f0' },
        horzLines: { color: '#f0f0f0' },
      },
      crosshair: {
        mode: 1,
      },
      rightPriceScale: {
        borderColor: '#ddd',
      },
      timeScale: {
        borderColor: '#ddd',
        timeVisible: true,
        secondsVisible: false,
      },
      height,
    })

    const candleSeries = chart.addCandlestickSeries({
      upColor: '#26a69a',
      downColor: '#ef5350',
      borderVisible: false,
      wickUpColor: '#26a69a',
      wickDownColor: '#ef5350',
    })

    chartRef.current = chart
    candleSeriesRef.current = candleSeries

    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({
          width: chartContainerRef.current.clientWidth,
        })
      }
    }

    window.addEventListener('resize', handleResize)
    handleResize()

    return () => {
      window.removeEventListener('resize', handleResize)
      chart.remove()
    }
  }, [height])

  useEffect(() => {
    if (!candleSeriesRef.current || data.length === 0) return

    const chartData: CandlestickData<Time>[] = data.map((candle) => ({
      time: Math.floor(new Date(candle.timestamp).getTime() / 1000) as Time,
      open: candle.open,
      high: candle.high,
      low: candle.low,
      close: candle.close,
    }))

    candleSeriesRef.current.setData(chartData)
    lastDataRef.current = data

    if (chartRef.current) {
      chartRef.current.timeScale().fitContent()
    }
  }, [data])

  useEffect(() => {
    if (!candleSeriesRef.current || lastPrice === undefined || data.length === 0) return

    const lastCandle = data[data.length - 1]
    const currentTime = Math.floor(Date.now() / 1000) as Time
    const candleTime = Math.floor(new Date(lastCandle.timestamp).getTime() / 1000) as Time

    const isSameCandle = (currentTime as number) - (candleTime as number) < 60

    if (isSameCandle) {
      const updatedCandle: CandlestickData<Time> = {
        time: candleTime,
        open: lastCandle.open,
        high: Math.max(lastCandle.high, lastPrice),
        low: Math.min(lastCandle.low, lastPrice),
        close: lastPrice,
      }
      candleSeriesRef.current.update(updatedCandle)
    } else {
      const newCandle: CandlestickData<Time> = {
        time: currentTime,
        open: lastPrice,
        high: lastPrice,
        low: lastPrice,
        close: lastPrice,
      }
      candleSeriesRef.current.update(newCandle)

      lastDataRef.current = [...lastDataRef.current, {
        timestamp: new Date().toISOString(),
        open: lastPrice,
        high: lastPrice,
        low: lastPrice,
        close: lastPrice,
      }]
    }
  }, [lastPrice, data])

  return (
    <div
      ref={chartContainerRef}
      style={{ width: '100%', height: `${height}px` }}
    />
  )
}