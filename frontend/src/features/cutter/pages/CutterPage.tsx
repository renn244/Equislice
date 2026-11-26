import { CheckCircleFillIcon } from '@primer/octicons-react'
import { SiteHeader } from '../../../components/site-header'
import { EmptyCutter } from '../components/EmptyCutter'
import { PanoramaJobPanel } from '../components/PanoramaJobPanel'
import { CutterForm } from '../forms/CutterForm'
import type { CutterFormValues } from '../forms/cutterForm.schema'
import { usePanoramaImage } from '../hooks/usePanoramaImage'
import { usePanoramaJob } from '../hooks/usePanoramaJob'
import './cutter-page.css'

export function CutterPage() {
  const image = usePanoramaImage()
  const job = usePanoramaJob()

  async function submit(values: CutterFormValues) {
    if (image.selectedFile) await job.submit(image.selectedFile, values)
  }

  return (
    <div className="cutter-page">
      <SiteHeader compact />
      <main className="cutter-main">
        {job.jobId && !image.selectedFile && (
          <div className="restored-session" role="status">
            <CheckCircleFillIcon aria-hidden="true" size={16} />
            <p><strong>Session restored.</strong> Checking the job saved on this device.</p>
            <code>{job.jobId}</code>
          </div>
        )}

        {image.selectedFile ? (
          <CutterForm
            selectedFile={image.selectedFile}
            dimensions={image.dimensions}
            previewUrl={image.previewUrl}
            fileError={image.fileError}
            jobState={job.jobState}
            onFileChange={image.onFileChange}
            onRemoveFile={image.reset}
            onSubmit={submit}
          />
        ) : (
          <EmptyCutter fileError={image.fileError} onFileChange={image.onFileChange} onSample={image.loadSample} />
        )}

        {(job.jobState !== 'idle' || job.jobError) && (
          <PanoramaJobPanel
            key={job.jobId ?? job.jobState}
            jobId={job.jobId}
            jobState={job.jobState}
            jobError={job.jobError}
            downloads={job.downloads}
            downloadColumns={job.downloadColumns}
          />
        )}
      </main>
    </div>
  )
}
