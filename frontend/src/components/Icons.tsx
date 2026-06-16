import React from 'react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import pluginsSvg from '../assets/icons/plugins.svg'; // Static import for UI icons

function cn(...inputs) {
 return twMerge(clsx(inputs));
}

const BaseIcon = ({ src, size = 20, className }) => (
 <div
 className={cn("bg-current inline-block shrink-0", className)}
 style={{
 width: size,
 height: size,
 maskImage: `url(${src})`,
 WebkitMaskImage: `url(${src})`,
 maskRepeat: 'no-repeat',
 WebkitMaskRepeat: 'no-repeat',
 maskPosition: 'center',
 WebkitMaskPosition: 'center',
 maskSize: 'contain',
 WebkitMaskSize: 'contain'
 }}
 />
);

const RawSVGIcon = ({ svgString, size = 20, className }) => {
 if (!svgString) return null;
 const cleanSvg = svgString.replace(/<\?xml.*\?>/g, "").trim();
 return (
 <div
 className={cn("bg-current inline-block shrink-0", className)}
 style={{
 width: size,
 height: size,
 maskImage: `url("data:image/svg+xml,${encodeURIComponent(cleanSvg)}")`,
 WebkitMaskImage: `url("data:image/svg+xml,${encodeURIComponent(cleanSvg)}")`,
 maskRepeat: 'no-repeat',
 WebkitMaskRepeat: 'no-repeat',
 maskPosition: 'center',
 WebkitMaskPosition: 'center',
 maskSize: 'contain',
 WebkitMaskSize: 'contain'
 }}
 />
 );
};

export const Plugins = (props) => <BaseIcon src={pluginsSvg} {...props} />;

const Icons = {
 Raw: RawSVGIcon,
 Plugins: Plugins
};

export default Icons;
