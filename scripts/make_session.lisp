#!/usr/local/bin/sbcl --script

#|
  Function handles the repetitive things about sessions in PDL
  Sets a environment variable as ticket
  Renames Tmux session to be ticket name
  Sees if a branch exist with ticket number
  If it does, checks out that branch
  If it does not, prompts for 'feature' 'bugfix' 'hotfix'
  Asks for description for new branch name
  creates a draft github pr 
  (with-open-file (stream "./.scratch/.session.info" :direction :output :if-exists :supersede)
    (format stream "CXPVSP-~a" name))
  )
|#
(defun set-ticket (name)
  "Sets a environment variable and stores it in a file located at ./.pdl/.session.info"
  (let ((ticket (concatenate 'string "CXPVSP-" (write-to-string name))))
    (ensure-directories-exist "~/.pdl/")
    (with-open-file (stream "~/.pdl/.session.info" :direction :output :if-exists :supersede)
      (format stream (concatenate 'string ticket ": " (ext:cd))))
    )
  )


(set-ticket 12)
