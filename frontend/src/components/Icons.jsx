import React from 'react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

// SVG Assets
import phpSvg from '../assets/icons/php.svg';
import apacheSvg from '../assets/icons/apache.svg';
import nginxSvg from '../assets/icons/nginx.svg';
import mysqlSvg from '../assets/icons/mysql.svg';
import nodeSvg from '../assets/icons/node.svg';
import heidisqlSvg from '../assets/icons/heidisql.svg';
import opensslSvg from '../assets/icons/openssl.svg';
import pluginsSvg from '../assets/icons/plugins.svg';

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

export const PHP = (props) => <BaseIcon src={phpSvg} {...props} />;
export const Apache = (props) => <BaseIcon src={apacheSvg} {...props} />;
export const Nginx = (props) => <BaseIcon src={nginxSvg} {...props} />;
export const MySQL = (props) => <BaseIcon src={mysqlSvg} {...props} />;
export const Node = (props) => <BaseIcon src={nodeSvg} {...props} />;
export const HeidiSQL = (props) => <BaseIcon src={heidisqlSvg} {...props} />;
export const OpenSSL = (props) => <BaseIcon src={opensslSvg} {...props} />;
export const Plugins = (props) => <BaseIcon src={pluginsSvg} {...props} />;

const Icons = {
  PHP,
  Apache,
  Nginx,
  MySQL,
  Node,
  HeidiSQL,
  OpenSSL,
  Plugins
};

export default Icons;
