import type { ButtonHTMLAttributes, ComponentType, HTMLAttributes, PropsWithChildren, ReactNode } from 'react'
import { LoaderCircle } from 'lucide-react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'icon'
}

export function Button({
  variant = 'secondary',
  size = 'md',
  className = '',
  children,
  ...props
}: PropsWithChildren<ButtonProps>) {
  return (
    <button className={`button button-${variant} button-${size} ${className}`} {...props}>
      {children}
    </button>
  )
}

export function Panel({ className = '', children, ...props }: PropsWithChildren<HTMLAttributes<HTMLDivElement>>) {
  return <section className={`panel ${className}`} {...props}>{children}</section>
}

export function PanelHeading({
  title,
  description,
  action,
}: {
  title: string
  description?: string
  action?: ReactNode
}) {
  return (
    <div className="panel-heading">
      <div className="min-w-0">
        <h2 className="panel-title">{title}</h2>
        {description && <p className="panel-description">{description}</p>}
      </div>
      {action}
    </div>
  )
}

export function EmptyState({
  icon: Icon,
  title,
  description,
}: {
  icon: ComponentType<{ size?: number; strokeWidth?: number }>
  title: string
  description: string
}) {
  return (
    <div className="empty-state">
      <div className="empty-icon"><Icon size={22} strokeWidth={1.7} /></div>
      <strong>{title}</strong>
      <p>{description}</p>
    </div>
  )
}

export function LoadingBlock({ label = 'Loading runtime data…' }: { label?: string }) {
  return <div className="loading-block"><LoaderCircle size={16} className="animate-spin" />{label}</div>
}

export function Eyebrow({ children }: PropsWithChildren) {
  return <div className="eyebrow">{children}</div>
}

export function IconButton({
  label,
  children,
  className = '',
  ...props
}: PropsWithChildren<ButtonHTMLAttributes<HTMLButtonElement> & { label: string }>) {
  return <button className={`icon-button ${className}`} aria-label={label} title={label} {...props}>{children}</button>
}
