import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import {
  ResourceLineChart,
  MetricItem,
  buildSparklineSegments,
  buildSparklinePaths,
} from '../../../src/components/ssh/ResourceLineChart';

describe('ResourceLineChart and MetricItem', () => {
  describe('buildSparklineSegments & buildSparklinePaths', () => {
    it('creates continuous segments for valid data', () => {
      const data = [
        { cpu: 10, mem: 20, disk: 30 },
        { cpu: 20, mem: 30, disk: 40 },
        { cpu: 30, mem: 40, disk: 50 },
      ];
      const segments = buildSparklineSegments(data, 'cpu', 100, 50, 3);
      expect(segments).toHaveLength(1);
      expect(segments[0]).toHaveLength(3);

      const paths = buildSparklinePaths(segments, 50);
      expect(paths.linePath).toContain('M 0');
      expect(paths.areaPath).toContain('M 0 50');
    });

    it('splits segments across null values', () => {
      const data = [
        { cpu: 10, mem: null, disk: null },
        { cpu: null, mem: null, disk: null },
        { cpu: 30, mem: null, disk: null },
      ];
      const segments = buildSparklineSegments(data, 'cpu', 100, 50, 3);
      expect(segments).toHaveLength(2);
      expect(segments[0]).toHaveLength(1);
      expect(segments[1]).toHaveLength(1);
    });
  });

  describe('ResourceLineChart component', () => {
    it('renders SVG without crashing with empty data', () => {
      const { container } = render(
        <ResourceLineChart
          data={[]}
          metric="cpu"
          color="#3b82f6"
          fillColor="rgba(59, 130, 246, 0.2)"
        />
      );
      expect(container.querySelector('svg')).toBeInTheDocument();
      expect(screen.getByRole('img', { name: 'CPU usage chart' })).toBeInTheDocument();
    });

    it('renders SVG paths when data is provided', () => {
      const data = [
        { cpu: 25, mem: 50, disk: 75 },
        { cpu: 50, mem: 60, disk: 80 },
      ];
      const { container } = render(
        <ResourceLineChart
          data={data}
          metric="cpu"
          color="#3b82f6"
          fillColor="rgba(59, 130, 246, 0.2)"
        />
      );
      expect(container.querySelectorAll('path').length).toBeGreaterThan(0);
    });
  });

  describe('MetricItem component', () => {
    it('renders label and value', () => {
      render(
        <MetricItem
          label="CPU"
          valueText="45%"
          metric="cpu"
          color="#3b82f6"
          fillColor="rgba(59, 130, 246, 0.2)"
          hoveredMetric={null}
          setHoveredMetric={vi.fn()}
          displayMode="always"
          history={[{ cpu: 45, mem: 50, disk: 50 }]}
        />
      );
      expect(screen.getByText('CPU: 45%')).toBeInTheDocument();
    });

    it('triggers mouseEnter and mouseLeave handlers', () => {
      const setHoveredMetric = vi.fn();
      render(
        <MetricItem
          label="Memory"
          valueText="60%"
          metric="mem"
          color="#10b981"
          fillColor="rgba(16, 185, 129, 0.2)"
          hoveredMetric={null}
          setHoveredMetric={setHoveredMetric}
          displayMode="tooltip"
          history={[]}
        />
      );

      const item = screen.getByText('Memory: 60%').parentElement!;
      fireEvent.mouseEnter(item);
      expect(setHoveredMetric).toHaveBeenCalledWith('mem');

      fireEvent.mouseLeave(item);
      expect(setHoveredMetric).toHaveBeenCalledWith(null);
    });
  });
});
