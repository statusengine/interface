<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Loader\Crate;

use Crate\PDO\PDO;
use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\ServiceAcknowledgementLoaderInterface;
use Statusengine\ValueObjects\ServiceAcknowledgementQueryOptions;

class ServiceAcknowledgementLoader implements ServiceAcknowledgementLoaderInterface {

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * ServiceAcknowledgementLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    /**
     * @param ServiceAcknowledgementQueryOptions $ServiceAcknowledgementQueryOptions
     * @return array
     */
    public function getAcknowledgements(ServiceAcknowledgementQueryOptions $ServiceAcknowledgementQueryOptions) {
        $fields = [
            'hostname',
            'service_description',
            'acknowledgement_type',
            'author_name',
            'comment_data',
            'entry_time',
            'is_sticky',
            'notify_contacts',
            'persistent_comment',
            'state'
        ];
        $baseQuery = sprintf(
            'SELECT %s FROM statusengine_service_acknowledgements'
            , implode(',', $fields)
        );

        if ($ServiceAcknowledgementQueryOptions->getHostname()) {
            $baseQuery = sprintf(' %s WHERE hostname=?', $baseQuery);
        }

        if ($ServiceAcknowledgementQueryOptions->getServiceDescription()) {
            if ($ServiceAcknowledgementQueryOptions->getHostname()) {
                $baseQuery = sprintf('%s AND service_description=?', $baseQuery);
            } else {
                $baseQuery = sprintf('%s WHERE service_description=?', $baseQuery);
            }
        }

        if ($ServiceAcknowledgementQueryOptions->sizeOfStateFilter() > 0 && $ServiceAcknowledgementQueryOptions->sizeOfStateFilter() < 4) {
            $baseQuery = sprintf('%s AND state IN(%s)', $baseQuery, implode(',', $ServiceAcknowledgementQueryOptions->getStateFilter()));
        }

        if ($ServiceAcknowledgementQueryOptions->getCommentDataLike() != '') {
            $baseQuery = sprintf(' %s AND comment_data ~* ? ', $baseQuery);
        }

        if ($ServiceAcknowledgementQueryOptions->getEntryTimeLt() > 0) {
            $baseQuery = sprintf(' %s AND entry_time < ? ', $baseQuery);
        }

        if ($ServiceAcknowledgementQueryOptions->getEntryTimeGt() > 0) {
            $baseQuery = sprintf(' %s AND entry_time > ? ', $baseQuery);
        }

        $baseQuery = sprintf(
            '%s ORDER BY %s %s LIMIT ? OFFSET ?',
            $baseQuery,
            $ServiceAcknowledgementQueryOptions->getOrder(),
            $ServiceAcknowledgementQueryOptions->getDirection()
        );


        $query = $this->Backend->prepare($baseQuery);

        $i = 1;
        if ($ServiceAcknowledgementQueryOptions->getHostname()) {
            $query->bindValue($i++, $ServiceAcknowledgementQueryOptions->getHostname());
        }

        if ($ServiceAcknowledgementQueryOptions->getServiceDescription()) {
            $query->bindValue($i++, $ServiceAcknowledgementQueryOptions->getServiceDescription());
        }


        if ($ServiceAcknowledgementQueryOptions->getCommentDataLike() != '') {
            $like = sprintf('.*%s.*', $ServiceAcknowledgementQueryOptions->getCommentDataLike());
            $query->bindValue($i++, $like);
        }

        if ($ServiceAcknowledgementQueryOptions->getEntryTimeLt() > 0) {
            $query->bindValue($i++, $ServiceAcknowledgementQueryOptions->getEntryTimeLt());
        }

        if ($ServiceAcknowledgementQueryOptions->getEntryTimeGt() > 0) {
            $query->bindValue($i++, $ServiceAcknowledgementQueryOptions->getEntryTimeGt());
        }

        $query->bindValue($i++, $ServiceAcknowledgementQueryOptions->getLimit(), PDO::PARAM_INT);
        $query->bindValue($i++, $ServiceAcknowledgementQueryOptions->getOffset(), PDO::PARAM_INT);

        return $this->Backend->fetchAll($query);

    }

    /**
     * @param ServiceAcknowledgementQueryOptions $ServiceAcknowledgementQueryOptions
     * @return array
     */
    public function getCurrentServiceAcknowledgements(ServiceAcknowledgementQueryOptions $ServiceAcknowledgementQueryOptions) {
        $baseQuery = sprintf('SELECT hostname, service_description, node_name from statusengine_servicestatus WHERE problem_has_been_acknowledged=true');
        if($ServiceAcknowledgementQueryOptions->getHostnameLike() != ''){
            $baseQuery = sprintf('%s AND hostname ~* ?', $baseQuery);
        }

        if($ServiceAcknowledgementQueryOptions->getServicedescriptionLike() != ''){
            $baseQuery = sprintf('%s AND service_description ~* ?', $baseQuery);
        }
        $baseQuery .= $this->getClusterNameQuery($ServiceAcknowledgementQueryOptions);

        $baseQuery = sprintf(' %s LIMIT ? OFFSET ?', $baseQuery);


        $query = $this->Backend->prepare($baseQuery);
        $i = 1;
        if ($ServiceAcknowledgementQueryOptions->getHostnameLike() != '') {
            $like = sprintf('.*%s.*', $ServiceAcknowledgementQueryOptions->getHostnameLike());
            $query->bindValue($i++, $like);
        }
        if ($ServiceAcknowledgementQueryOptions->getServicedescriptionLike() != '') {
            $like = sprintf('.*%s.*', $ServiceAcknowledgementQueryOptions->getServicedescriptionLike());
            $query->bindValue($i++, $like);
        }

        foreach ($ServiceAcknowledgementQueryOptions->getClusterName() as $clusterName) {
            $query->bindValue($i++, $clusterName);
        }

        $query->bindValue($i++, $ServiceAcknowledgementQueryOptions->getLimit(), PDO::PARAM_INT);
        $query->bindValue($i++, $ServiceAcknowledgementQueryOptions->getOffset(), PDO::PARAM_INT);

        $result = $this->Backend->fetchAll($query);

        $mergedResult = [];
        foreach($result as $row){
            $baseQuery = 'SELECT * FROM statusengine_service_acknowledgements WHERE hostname=? AND service_description=? ORDER BY entry_time DESC LIMIT 1';
            $ackQuery = $this->Backend->prepare($baseQuery);
            $ackQuery->bindParam(1, $row['hostname']);
            $ackQuery->bindParam(2, $row['service_description']);
            $ackResult = $this->Backend->fetchAll($ackQuery);
            foreach($ackResult as $record){
                $mergedResult[] = array_merge($record, $row);
            }
        }

        return $mergedResult;
    }

    /**
     * @param ServiceAcknowledgementQueryOptions $QueryOptions
     * @return string
     */
    private function getClusterNameQuery(ServiceAcknowledgementQueryOptions $QueryOptions) {
        $placeholders = [];
        foreach ($QueryOptions->getClusterName() as $clusterName) {
            $placeholders[] = '?';
        }
        if (!empty($placeholders)) {
            return sprintf(' AND node_name IN(%s)', implode(',', $placeholders));
        }
        return '';
    }

}
