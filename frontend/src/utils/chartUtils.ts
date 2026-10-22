export type ResourceMetricType = "cpu" | "mem" | "disk";

export interface ResourceDataPoint {
  cpu: number | null;
  mem: number | null;
  disk: number | null;
}

export interface SparklinePaths {
  linePath: string;
  areaPath: string;
}

/**
 * Calculates continuous coordinate segments for valid numeric points,
 * breaking whenever a null/missing data point is encountered.
 */
export const buildSparklineSegments = (
  paddedData: ResourceDataPoint[],
  metric: ResourceMetricType,
  width: number,
  height: number,
  pointsCount: number,
): Array<Array<[number, number]>> => {
  const getX = (index: number) => (index / (pointsCount - 1)) * width;
  const getY = (val: number | null) =>
    val === null ? height : height - 1 - (val / 100) * (height - 2);

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

/**
 * Generates SVG path commands (M, L, Z) for the stroke line and filled area.
 */
export const buildSparklinePaths = (
  segments: Array<Array<[number, number]>>,
  height: number,
): SparklinePaths => {
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
