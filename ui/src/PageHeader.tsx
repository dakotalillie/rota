type Props = {
  title: string;
  actions?: React.ReactNode;
};

export default function PageHeader({ title, actions }: Props) {
  return (
    <div className="flex items-center justify-between">
      <h1 className="text-xl font-bold tracking-tight">{title}</h1>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  );
}
