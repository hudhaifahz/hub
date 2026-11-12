import React from "react";
import { Separator } from "src/components/ui/separator";
import { GRYNVAULT_PRODUCT_NAME } from "src/lib/grynvault";

type Props = {
  title: string;
  description: string | React.ReactNode;
  pageTitle?: string;
};

function SettingsHeader({ title, description, pageTitle }: Props) {
  return (
    <>
      {pageTitle && <title>{`${pageTitle} · ${GRYNVAULT_PRODUCT_NAME}`}</title>}
      <div className="space-y-6">
        <div>
          <h3 className="text-2xl font-medium">{title}</h3>
          <p className="text-base text-muted-foreground">{description}</p>
        </div>
        <Separator />
      </div>
    </>
  );
}

export default SettingsHeader;
