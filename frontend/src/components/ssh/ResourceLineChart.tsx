import React from "react";

export type ResourceMetricType = "cpu" | "mem" | "disk";

export interface ResourceDataPoint {
  cpu: number | null;
  mem: number | null;
  disk: number | null;
}

export interface ResourceLineChartProps {
  /** Sequential historical data of resource metrics containing CPU, Memory, and Disk ratios */
  data: ResourceDataPoint[];
  /** Metric type to parse and plot */
  metric: ResourceMetricType;
  /** Stroke color of the plotted sparkline */
  color: string;
  /** Background area gradient or fill color of the sparkline */
  fillColor: string;
}

export const buildSparklineSegments = (
  paddedData: ResourceDataPoint[],
  metric: ResourceMetricType,
  width: number,
  height: number,
  pointsCount: number
) => {
  const getX = (index: number) => (index / (pointsCount - 1)) * width;
  const getY = (val: number | null) => (val === null ? height : height - 1 - (val / 100) * (height - 2));

  let currentSegment: Array<[number, number]> = [];
  const segments: Array<Array<[number, number]>> = [];

  paddedData.forEach((d, i) => {
    const val = d[metric];
    if (val !== null) {
      currentSegment.push([getX(i), getY(val)]);
    } else if (currentSegment.length > 0) {
      segments.push(currentSegment);
      currentSegment = [];
    }
  });
  if (currentSegment.length > 0) {
    segments.push(currentSegment);
  }
  return segments;
};

export const buildSparklinePaths = (segments: Array<Array<[number, number]>>, height: number) => {
  let linePath = "";
  let areaPath = "";

  segments.forEach((seg) => {
    if (seg.length === 0) return;

    let segLine = `M ${seg[0][0]} ${seg[0][1]}`;
    let segArea = `M ${seg[0][0]} ${height} L ${seg[0][0]} ${seg[0][1]}`;

    for (let j = 1; j < seg.length; j++) {
      segLine += ` L ${seg[j][0]} ${seg[j][1]}`;
      segArea += ` L ${seg[j][0]} ${seg[j][1]}`;
    }

    segArea += ` L ${seg.at(-1)![0]} ${height} Z`;

    linePath += " " + segLine;
    areaPath += " " + segArea;
  });

  return { linePath, areaPath };
};

export const ResourceLineChart: React.FC<ResourceLineChartProps> = ({ data, metric, color, fillColor }) => {
  const width = 120;
  const height = 30;
  const pointsCount = 30;

  const paddedData = [...new Array(pointsCount).fill({ cpu: null, mem: null, disk: null }), ...data].slice(-pointsCount);
  const segments = buildSparklineSegments(paddedData, metric, width, height, pointsCount);
  const { linePath, areaPath } = buildSparklinePaths(segments, height);

  const chartTitle = `${metric.toUpperCase()} usage chart`;

  return (
    <svg
      role="img"
      aria-label={chartTitle}
      width={width}
      height={height}
      className="overflow-hidden border border-mui-grey-200 dark:border-white/10 rounded bg-white dark:bg-mui-grey-950"
    >
      <title>{chartTitle}</title>
      {/* Horizontal grid lines */}
      <line x1={0} y1={height / 3} x2={width} y2={height / 3} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />
      <line x1={0} y1={(2 * height) / 3} x2={width} y2={(2 * height) / 3} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />

      {/* Vertical grid lines */}
      <line x1={width / 4} y1={0} x2={width / 4} y2={height} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />
      <line x1={width / 2} y1={0} x2={width / 2} y2={height} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />
      <line x1={(3 * width) / 4} y1={0} x2={(3 * width) / 4} y2={height} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />

      {areaPath && <path d={areaPath} fill={fillColor} />}
      {linePath && <path d={linePath} fill="none" stroke={color} strokeWidth={1.5} />}
    </svg>
  );
};

export interface MetricItemProps {
  label: string;
  valueText: string;
  metric: ResourceMetricType;
  color: string;
  fillColor: string;
  hoveredMetric: ResourceMetricType | null;
  setHoveredMetric: (m: ResourceMetricType | null) => void;
  displayMode: string;
  history: ResourceDataPoint[];
  minWidthClass?: string;
}

export const MetricItem: React.FC<MetricItemProps> = ({
  label,
  valueText,
  metric,
  color,
  fillColor,
  hoveredMetric,
  setHoveredMetric,
  displayMode,
  history,
  minWidthClass = "min-w-[55px]",
}) => {
  const isHovered = hoveredMetric === metric;
  const showTooltip = displayMode === "tooltip" && isHovered;
  const showInline =
    displayMode !== "tooltip" &&
    (displayMode === "always" || (displayMode === "hover-inline" && isHovered));

  return (
    <div
      className="relative py-1 cursor-help group flex items-center gap-2"
      onMouseEnter={() => setHoveredMetric(metric)}
      onMouseLeave={() => setHoveredMetric(null)}
    >
      <span className={minWidthClass}>
        {label}: {valueText}
      </span>
      {showTooltip && (
        <div className="absolute bottom-7 left-0 z-50 bg-white dark:bg-mui-grey-850 p-2 rounded shadow-lg border border-mui-grey-200 dark:border-white/10 flex flex-col gap-1 items-center animate-in fade-in duration-100 min-w-[130px]">
          <span className="text-[9px] uppercase tracking-wider text-mui-grey-500 dark:text-mui-grey-400">
            {label} Usage History
          </span>
          <ResourceLineChart
            data={history}
            metric={metric}
            color={color}
            fillColor={fillColor}
          />
        </div>
      )}
      {showInline && (
        <div className="flex items-center shrink-0">
          <ResourceLineChart
            data={history}
            metric={metric}
            color={color}
            fillColor={fillColor}
          />
        </div>
      )}
    </div>
  );
};
